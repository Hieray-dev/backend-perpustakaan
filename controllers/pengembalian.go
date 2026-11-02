package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"time"
)

type PengembalianRequest struct {
	IDPeminjaman int `json:"id_peminjaman"`
	IDPetugas    int `json:"id_petugas"`
}

type Pengembalian struct {
	IDPengembalian      int    `json:"id_pengembalian"`
	IDPeminjaman        int    `json:"id_peminjaman"`
	KodeTransaksi       string `json:"kode_transaksi"`
	JudulBuku           string `json:"judul_buku"`
	Peminjam            string `json:"peminjam"`
	Petugas             string `json:"petugas"`
	TanggalPeminjaman   string `json:"tanggal_peminjaman"`
	TanggalKembali      string `json:"tanggal_kembali"`
	TanggalPengembalian string `json:"tanggal_pengembalian"`
	Denda               int    `json:"denda"`
}

func GetPengembalianHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := config.DB.Query(`
		SELECT pg.id_pengembalian, COALESCE(pg.id_peminjaman, 0),
		       COALESCE(p.kode_transaksi, ''), COALESCE(b.judul, ''),
		       COALESCE(up.username, ''), COALESCE(ut.username, ''),
		       COALESCE(p.tanggal_peminjaman, ''), COALESCE(p.tanggal_kembali, ''),
		       pg.tanggal_pengembalian, COALESCE(pg.denda, 0)
		FROM pengembalian pg
		LEFT JOIN peminjaman p ON pg.id_peminjaman = p.id_peminjaman
		LEFT JOIN buku b ON p.id_buku = b.id_buku
		LEFT JOIN user up ON p.id_peminjam = up.id_user
		LEFT JOIN user ut ON p.id_petugas = ut.id_user
		ORDER BY pg.id_pengembalian DESC`)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengambil data pengembalian: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listPengembalian []Pengembalian
	for rows.Next() {
		var pg Pengembalian
		err := rows.Scan(
			&pg.IDPengembalian, &pg.IDPeminjaman,
			&pg.KodeTransaksi, &pg.JudulBuku,
			&pg.Peminjam, &pg.Petugas,
			&pg.TanggalPeminjaman, &pg.TanggalKembali,
			&pg.TanggalPengembalian, &pg.Denda,
		)
		if err != nil {
			http.Error(w, `{"message":"Gagal membaca data pengembalian"}`, http.StatusInternalServerError)
			return
		}
		listPengembalian = append(listPengembalian, pg)
	}

	json.NewEncoder(w).Encode(listPengembalian)
}

func PengembalianHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req PengembalianRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	var tanggalKembali sql.NullString
	var idBuku int
	var statusSaatIni string

	queryCek := `SELECT tanggal_kembali, id_buku, status FROM peminjaman WHERE id_peminjaman = ?`
	err = config.DB.QueryRow(queryCek, req.IDPeminjaman).Scan(&tanggalKembali, &idBuku, &statusSaatIni)
	if err != nil {
		http.Error(w, "Data peminjaman tidak ditemukan", http.StatusNotFound)
		return
	}

	if statusSaatIni != "dipinjam" {
		http.Error(w, "Buku tidak dalam status dipinjam (mungkin masih booking atau sudah dikembalikan)", http.StatusBadRequest)
		return
	}

	waktuSekarang := time.Now()
	tanggalSekarangStr := waktuSekarang.Format("2006-01-02")

	var tKembali time.Time
	if tanggalKembali.Valid && len(tanggalKembali.String) >= 10 {
		tKembali, _ = time.Parse("2006-01-02", tanggalKembali.String[:10])
	}
	tSekarang, _ := time.Parse("2006-01-02", tanggalSekarangStr)

	denda := 0
	if !tKembali.IsZero() && tSekarang.After(tKembali) {
		selisihHari := int(tSekarang.Sub(tKembali).Hours() / 24)
		if selisihHari > 0 {
			denda = selisihHari * 2000
		}
	}

	queryInsertPengembalian := `
		INSERT INTO pengembalian (tanggal_pengembalian, denda, id_peminjaman)
		VALUES (?, ?, ?)`
	_, err = config.DB.Exec(queryInsertPengembalian, tanggalSekarangStr, denda, req.IDPeminjaman)
	if err != nil {
		http.Error(w, "Gagal mencatat pengembalian: "+err.Error(), http.StatusInternalServerError)
		return
	}

	queryUpdatePinjam := `
		UPDATE peminjaman 
		SET status = 'dikembalikan', tanggal_kembali = ?, id_petugas = ? 
		WHERE id_peminjaman = ?`
	config.DB.Exec(queryUpdatePinjam, tanggalSekarangStr, req.IDPetugas, req.IDPeminjaman)

	queryUpdateStok := `UPDATE buku SET stok = stok + 1 WHERE id_buku = ?`
	config.DB.Exec(queryUpdateStok, idBuku)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":              "Buku berhasil dikembalikan!",
		"id_peminjaman":        req.IDPeminjaman,
		"tanggal_pengembalian": tanggalSekarangStr,
		"denda":                denda,
		"status_transaksi":     "dikembalikan",
	})
}

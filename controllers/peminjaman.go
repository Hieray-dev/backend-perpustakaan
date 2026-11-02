package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/middleware"
	"perpustakaan-api/models"
	"perpustakaan-api/policy"
	"perpustakaan-api/utils"
	"time"
)

func GetRiwayatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.UserCtxKey).(*utils.Claims)
	if !ok || claims == nil {
		http.Error(w, `{"message":"Tidak terautentikasi"}`, http.StatusUnauthorized)
		return
	}

	baseQuery := `
	SELECT
		p.id_peminjaman,
		p.kode_transaksi,
		b.judul,
		COALESCE(p.id_peminjam, 0),
		COALESCE(up.username, ''),
		p.tanggal_peminjaman,
		p.status,
		COALESCE(u.username, 'Belum Diverifikasi') AS pustakawan,
		COALESCE(u.is_active, 0) AS is_active
	FROM peminjaman p
	JOIN buku b ON p.id_buku = b.id_buku
	LEFT JOIN user up ON p.id_peminjam = up.id_user
	LEFT JOIN user u ON p.id_petugas = u.id_user`

	var (
		rows *sql.Rows
		err  error
	)
	if claims.IDRole == policy.RoleAnggota {
		rows, err = config.DB.Query(baseQuery+" WHERE p.id_peminjam = ? ORDER BY p.id_peminjaman DESC", claims.IDUser)
	} else {
		rows, err = config.DB.Query(baseQuery + " ORDER BY p.id_peminjaman DESC")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listRiwayat []models.Riwayat
	for rows.Next() {
		var rwt models.Riwayat
		var isActive int
		err := rows.Scan(
			&rwt.IDPeminjaman,
			&rwt.KodeTransaksi,
			&rwt.JudulBuku,
			&rwt.IDPeminjam,
			&rwt.Peminjam,
			&rwt.TanggalPinjam,
			&rwt.Status,
			&rwt.Pustakawan,
			&isActive,
		)
		if err != nil {
			continue
		}

		if rwt.Pustakawan == "Belum Diverifikasi" {
			rwt.StatusPetugas = "Menunggu Verifikasi"
		} else if isActive == 1 {
			rwt.StatusPetugas = "Aktif"
		} else {
			rwt.StatusPetugas = "Nonaktif (Resign)"
		}

		listRiwayat = append(listRiwayat, rwt)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listRiwayat)
}

func CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req models.BookingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	var stok int
	err = config.DB.QueryRow("SELECT stok FROM buku WHERE id_buku = ?", req.IDBuku).Scan(&stok)
	if err != nil || stok <= 0 {
		http.Error(w, `{"message":"Stok buku habis atau tidak ditemukan"}`, http.StatusBadRequest)
		return
	}

	kode := "BK-" + time.Now().Format("0201150405")
	batasPengambilan := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	queryInsert := `INSERT INTO peminjaman (kode_transaksi, tipe, tanggal_peminjaman, batas_waktu, status, id_peminjam, id_buku)
	VALUES (?, 'booking', CURRENT_DATE, ?, 'booking', ?, ?)`

	_, err = config.DB.Exec(queryInsert, kode, batasPengambilan, req.IDUser, req.IDBuku)
	if err != nil {
		http.Error(w, `{"message":"Gagal membuat booking: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	config.DB.Exec("UPDATE buku SET stok = stok - 1 WHERE id_buku = ?", req.IDBuku)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "Booking berhasil",
		"kode_transaksi":    kode,
		"batas_waktu_ambil": batasPengambilan,
	})
}

func ConfirmPinjam(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req models.ConfirmPinjamRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	tanggalPinjam := time.Now().Format("2006-01-02")
	tanggalKembali := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	queryUpdate := `UPDATE peminjaman
					SET status = 'dipinjam',
						tanggal_peminjaman = ?,
						tanggal_kembali = ?,
						id_petugas = ?
					WHERE id_peminjaman = ? AND status = 'booking'`

	res, err := config.DB.Exec(queryUpdate, tanggalPinjam, tanggalKembali, req.IDPetugas, req.IDPeminjaman)
	if err != nil {
		http.Error(w, `{"message": "Gagal memproses peminjaman: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"message":"Data booking tidak ditemukan atau sudah diproses"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "Buku berhasil diserahkan ke peminjam",
		"tanggal_peminjam": tanggalPinjam,
		"tanggal_kembali":  tanggalKembali,
	})
}

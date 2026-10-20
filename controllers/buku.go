package controllers

import (
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/models"
	"strconv"
	"strings"
)

func GetKatalogBuku(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := config.DB.Query(`
		SELECT b.id_buku, b.judul, COALESCE(b.penulis, ''), COALESCE(b.tahun_terbit, 0),
		       COALESCE(b.deskripsi, ''), COALESCE(b.gambar, ''), b.stok,
		       COALESCE(b.id_penerbit, 0), COALESCE(p.nama_penerbit, ''),
		       COALESCE(b.id_genre, 0), COALESCE(g.nama_genre, '')
		FROM buku b
		LEFT JOIN penerbit p ON b.id_penerbit = p.id_penerbit
		LEFT JOIN genre g ON b.id_genre = g.id_genre
		ORDER BY b.id_buku`)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengambil data buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listBuku []models.Buku
	for rows.Next() {
		var b models.Buku
		err := rows.Scan(&b.IDBuku, &b.Judul, &b.Penulis, &b.TahunTerbit, &b.Deskripsi, &b.Gambar, &b.Stok,
			&b.IDPenerbit, &b.Penerbit, &b.IDGenre, &b.Genre)
		if err != nil {
			http.Error(w, `{"message":"Gagal membaca data buku"}`, http.StatusInternalServerError)
			return
		}
		listBuku = append(listBuku, b)
	}

	json.NewEncoder(w).Encode(listBuku)
}

func CreateBukuHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var b models.Buku
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	// kolom yang boleh kosong: angka 0 dari request dianggap "tidak diisi" lalu disimpan NULL
	tahun := kosongkanNol(b.TahunTerbit)
	idPenerbit := kosongkanNol(b.IDPenerbit)
	idGenre := kosongkanNol(b.IDGenre)

	query := `INSERT INTO buku (judul, penulis, tahun_terbit, deskripsi, gambar, stok, id_penerbit, id_genre) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = config.DB.Exec(query, b.Judul, b.Penulis, tahun, b.Deskripsi, b.Gambar, b.Stok, idPenerbit, idGenre)
	if err != nil {
		http.Error(w, `{"message":"Gagal menambah buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Buku berhasil ditambahkan!",
	})
}

// kosongkanNol mengubah angka 0 menjadi NULL sebelum disimpan ke database.
// Dipakai untuk kolom yang boleh kosong: tahun_terbit, id_penerbit, id_genre.
// Tanpa ini, angka 0 ikut dikirim ke kolom relasi dan ditolak oleh foreign key.
func kosongkanNol(nilai int) interface{} {
	if nilai > 0 {
		return nilai
	}
	return nil
}

// updateBukuRequest memakai pointer supaya bisa membedakan:
// field tidak dikirim (nil) dan field dikirim dengan nilai 0/kosong.
type updateBukuRequest struct {
	Judul       *string `json:"judul"`
	Penulis     *string `json:"penulis"`
	TahunTerbit *int    `json:"tahun_terbit"`
	Deskripsi   *string `json:"deskripsi"`
	Gambar      *string `json:"gambar"`
	Stok        *int    `json:"stok"`
	IDPenerbit  *int    `json:"id_penerbit"`
	IDGenre     *int    `json:"id_genre"`
	IDStatus    *int    `json:"id_status"`
}

// UpdateBukuHandler mengubah sebagian field buku. Hanya field yang dikirim
// yang diubah; field yang tidak dikirim tetap seperti semula.
func UpdateBukuHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idBuku, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idBuku <= 0 {
		http.Error(w, `{"message":"ID buku tidak valid"}`, http.StatusBadRequest)
		return
	}

	var req updateBukuRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	// penjaga sederhana sebelum validasi lengkap (item 21)
	if req.Judul != nil && *req.Judul == "" {
		http.Error(w, `{"message":"Judul tidak boleh kosong"}`, http.StatusBadRequest)
		return
	}
	if req.Stok != nil && *req.Stok < 0 {
		http.Error(w, `{"message":"Stok tidak boleh negatif"}`, http.StatusBadRequest)
		return
	}
	if req.TahunTerbit != nil && *req.TahunTerbit < 0 {
		http.Error(w, `{"message":"Tahun terbit tidak valid"}`, http.StatusBadRequest)
		return
	}
	if req.IDPenerbit != nil && *req.IDPenerbit < 0 {
		http.Error(w, `{"message":"ID penerbit tidak valid"}`, http.StatusBadRequest)
		return
	}
	if req.IDGenre != nil && *req.IDGenre < 0 {
		http.Error(w, `{"message":"ID genre tidak valid"}`, http.StatusBadRequest)
		return
	}
	if req.IDStatus != nil && *req.IDStatus < 0 {
		http.Error(w, `{"message":"ID status tidak valid"}`, http.StatusBadRequest)
		return
	}

	bagianSet := []string{}
	argumen := []interface{}{}

	if req.Judul != nil {
		bagianSet = append(bagianSet, "judul = ?")
		argumen = append(argumen, *req.Judul)
	}
	if req.Penulis != nil {
		bagianSet = append(bagianSet, "penulis = ?")
		argumen = append(argumen, *req.Penulis)
	}
	if req.TahunTerbit != nil {
		bagianSet = append(bagianSet, "tahun_terbit = ?")
		argumen = append(argumen, kosongkanNol(*req.TahunTerbit))
	}
	if req.Deskripsi != nil {
		bagianSet = append(bagianSet, "deskripsi = ?")
		argumen = append(argumen, *req.Deskripsi)
	}
	if req.Gambar != nil {
		bagianSet = append(bagianSet, "gambar = ?")
		argumen = append(argumen, *req.Gambar)
	}
	if req.Stok != nil {
		bagianSet = append(bagianSet, "stok = ?")
		argumen = append(argumen, *req.Stok)
	}
	if req.IDPenerbit != nil {
		bagianSet = append(bagianSet, "id_penerbit = ?")
		argumen = append(argumen, kosongkanNol(*req.IDPenerbit))
	}
	if req.IDGenre != nil {
		bagianSet = append(bagianSet, "id_genre = ?")
		argumen = append(argumen, kosongkanNol(*req.IDGenre))
	}
	if req.IDStatus != nil {
		bagianSet = append(bagianSet, "id_status = ?")
		argumen = append(argumen, kosongkanNol(*req.IDStatus))
	}

	if len(bagianSet) == 0 {
		http.Error(w, `{"message":"Tidak ada data yang diubah"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM buku WHERE id_buku = ?", idBuku).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Buku tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	query := "UPDATE buku SET " + strings.Join(bagianSet, ", ") + " WHERE id_buku = ?"
	argumen = append(argumen, idBuku)
	_, err = config.DB.Exec(query, argumen...)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengubah buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Buku berhasil diubah!",
	})
}

// DeleteBukuHandler menghapus buku. Buku yang sedang dibooking, dipinjam,
// atau terlambat tidak boleh dihapus supaya data peminjaman tidak menggantung.
func DeleteBukuHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idBuku, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idBuku <= 0 {
		http.Error(w, `{"message":"ID buku tidak valid"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM buku WHERE id_buku = ?", idBuku).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Buku tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	var sedangDipinjam int
	err = config.DB.QueryRow(
		"SELECT COUNT(*) FROM peminjaman WHERE id_buku = ? AND status IN ('booking', 'dipinjam', 'terlambat')",
		idBuku,
	).Scan(&sedangDipinjam)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa peminjaman: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if sedangDipinjam > 0 {
		http.Error(w, `{"message":"Buku sedang dibooking atau dipinjam, tidak bisa dihapus"}`, http.StatusConflict)
		return
	}

	_, err = config.DB.Exec("DELETE FROM buku WHERE id_buku = ?", idBuku)
	if err != nil {
		http.Error(w, `{"message":"Gagal menghapus buku: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Buku berhasil dihapus!",
	})
}

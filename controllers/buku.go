package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

	b.Judul = strings.TrimSpace(b.Judul)
	if b.Judul == "" {
		http.Error(w, `{"message":"Judul wajib diisi"}`, http.StatusBadRequest)
		return
	}
	if b.Stok < 0 {
		http.Error(w, `{"message":"Stok tidak boleh negatif"}`, http.StatusBadRequest)
		return
	}
	if pesan, kode, ok := periksaRelasiBuku(b.IDPenerbit, b.IDGenre, 0); !ok {
		http.Error(w, `{"message":"`+pesan+`"}`, kode)
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

// relasiAda memeriksa apakah id yang diisi benar-benar ada di tabelnya.
// Nama tabel dan kolom selalu ditulis tetap di kode, bukan dari input pengguna.
func relasiAda(tabel, kolom string, id int) (bool, error) {
	var jumlah int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM "+tabel+" WHERE "+kolom+" = ?", id).Scan(&jumlah)
	return jumlah > 0, err
}

// periksaRelasiBuku memeriksa id_penerbit, id_genre, dan id_status yang diisi.
// Angka 0 berarti NULL, jadi tidak diperiksa. Mengembalikan pesan, kode HTTP,
// dan status lolos.
func periksaRelasiBuku(idPenerbit, idGenre, idStatus int) (string, int, bool) {
	if idPenerbit > 0 {
		ada, err := relasiAda("penerbit", "id_penerbit", idPenerbit)
		if err != nil {
			return "Gagal memeriksa penerbit: " + err.Error(), http.StatusInternalServerError, false
		}
		if !ada {
			return "Penerbit tidak ditemukan", http.StatusBadRequest, false
		}
	}
	if idGenre > 0 {
		ada, err := relasiAda("genre", "id_genre", idGenre)
		if err != nil {
			return "Gagal memeriksa genre: " + err.Error(), http.StatusInternalServerError, false
		}
		if !ada {
			return "Genre tidak ditemukan", http.StatusBadRequest, false
		}
	}
	if idStatus > 0 {
		ada, err := relasiAda("`status`", "id_status", idStatus)
		if err != nil {
			return "Gagal memeriksa status: " + err.Error(), http.StatusInternalServerError, false
		}
		if !ada {
			return "Status tidak ditemukan", http.StatusBadRequest, false
		}
	}
	return "", 0, true
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
	if req.Judul != nil {
		judul := strings.TrimSpace(*req.Judul)
		if judul == "" {
			http.Error(w, `{"message":"Judul tidak boleh kosong"}`, http.StatusBadRequest)
			return
		}
		req.Judul = &judul
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

	var cekPenerbit, cekGenre, cekStatus int
	if req.IDPenerbit != nil {
		cekPenerbit = *req.IDPenerbit
	}
	if req.IDGenre != nil {
		cekGenre = *req.IDGenre
	}
	if req.IDStatus != nil {
		cekStatus = *req.IDStatus
	}
	if pesan, kode, ok := periksaRelasiBuku(cekPenerbit, cekGenre, cekStatus); !ok {
		http.Error(w, `{"message":"`+pesan+`"}`, kode)
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

const batasUkuranGambar = 2 << 20 // 2 MB

// tipeGambarDiizinkan memetakan tipe isi berkas (hasil deteksi, bukan ekstensi
// nama) ke ekstensi yang dipakai saat menyimpan.
var tipeGambarDiizinkan = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadGambarBukuHandler menerima multipart/form-data (field "gambar"),
// menyimpan berkas dengan nama acak di ./uploads, lalu mengembalikan URL-nya.
// Kolom buku.gambar diisi terpisah lewat PUT /buku/{id}.
func UploadGambarBukuHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	// batas kasar di level pembacaan; batas pasti dicek lewat header.Size
	r.Body = http.MaxBytesReader(w, r.Body, batasUkuranGambar+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		http.Error(w, `{"message":"Gambar terlalu besar atau format multipart tidak valid"}`, http.StatusBadRequest)
		return
	}

	berkas, header, err := r.FormFile("gambar")
	if err != nil {
		http.Error(w, `{"message":"Field gambar wajib diisi"}`, http.StatusBadRequest)
		return
	}
	defer berkas.Close()

	if header.Size > batasUkuranGambar {
		http.Error(w, `{"message":"Ukuran gambar maksimal 2 MB"}`, http.StatusBadRequest)
		return
	}

	ekstensiNama := strings.ToLower(filepath.Ext(header.Filename))
	if ekstensiNama != ".jpg" && ekstensiNama != ".jpeg" && ekstensiNama != ".png" && ekstensiNama != ".webp" {
		http.Error(w, `{"message":"Format gambar harus jpg, png, atau webp"}`, http.StatusBadRequest)
		return
	}

	// periksa isi berkas, bukan hanya nama. Berkas teks yang dinamai .png tetap ditolak.
	kepala := make([]byte, 512)
	jumlahBaca, err := berkas.Read(kepala)
	if err != nil && err != io.EOF {
		http.Error(w, `{"message":"Gagal membaca gambar"}`, http.StatusInternalServerError)
		return
	}

	ekstensi, diizinkan := tipeGambarDiizinkan[http.DetectContentType(kepala[:jumlahBaca])]
	if !diizinkan {
		http.Error(w, `{"message":"Isi berkas bukan gambar jpg, png, atau webp"}`, http.StatusBadRequest)
		return
	}

	if _, err := berkas.Seek(0, io.SeekStart); err != nil {
		http.Error(w, `{"message":"Gagal membaca gambar"}`, http.StatusInternalServerError)
		return
	}

	acak := make([]byte, 16)
	if _, err := rand.Read(acak); err != nil {
		http.Error(w, `{"message":"Gagal membuat nama berkas"}`, http.StatusInternalServerError)
		return
	}
	namaBerkas := hex.EncodeToString(acak) + ekstensi

	if err := os.MkdirAll("./uploads", 0755); err != nil {
		http.Error(w, `{"message":"Gagal menyiapkan folder upload"}`, http.StatusInternalServerError)
		return
	}

	tujuan := filepath.Join("./uploads", namaBerkas)
	berkasTujuan, err := os.Create(tujuan)
	if err != nil {
		http.Error(w, `{"message":"Gagal menyimpan gambar"}`, http.StatusInternalServerError)
		return
	}
	defer berkasTujuan.Close()

	if _, err := io.Copy(berkasTujuan, berkas); err != nil {
		os.Remove(tujuan)
		http.Error(w, `{"message":"Gagal menyimpan gambar"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"url": "/uploads/" + namaBerkas,
	})
}

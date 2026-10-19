package controllers

import (
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/models"
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

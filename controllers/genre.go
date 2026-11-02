package controllers

import (
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"strconv"
	"strings"
)

type Genre struct {
	IDGenre   int    `json:"id_genre"`
	NamaGenre string `json:"nama_genre"`
}

func GetGenreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	rows, err := config.DB.Query("SELECT id_genre, nama_genre FROM genre")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listGenre []Genre
	for rows.Next() {
		var g Genre
		if err := rows.Scan(&g.IDGenre, &g.NamaGenre); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		listGenre = append(listGenre, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listGenre)
}

func CreateGenreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var g Genre
	err := json.NewDecoder(r.Body).Decode(&g)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	g.NamaGenre = strings.TrimSpace(g.NamaGenre)
	if g.NamaGenre == "" {
		http.Error(w, `{"message":"Nama genre wajib diisi"}`, http.StatusBadRequest)
		return
	}

	query := "INSERT INTO genre (nama_genre) VALUES (?)"
	_, err = config.DB.Exec(query, g.NamaGenre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Genre berhasil ditambahkan!",
	})
}

func UpdateGenreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idGenre, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idGenre <= 0 {
		http.Error(w, `{"message":"ID genre tidak valid"}`, http.StatusBadRequest)
		return
	}

	var g Genre
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	g.NamaGenre = strings.TrimSpace(g.NamaGenre)
	if g.NamaGenre == "" {
		http.Error(w, `{"message":"Nama genre tidak boleh kosong"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM genre WHERE id_genre = ?", idGenre).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa genre: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Genre tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("UPDATE genre SET nama_genre = ? WHERE id_genre = ?", g.NamaGenre, idGenre)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengubah genre: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Genre berhasil diubah!",
	})
}

func DeleteGenreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idGenre, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idGenre <= 0 {
		http.Error(w, `{"message":"ID genre tidak valid"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM genre WHERE id_genre = ?", idGenre).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa genre: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Genre tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("DELETE FROM genre WHERE id_genre = ?", idGenre)
	if err != nil {
		http.Error(w, `{"message":"Gagal menghapus genre: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Genre berhasil dihapus!",
	})
}

package controllers

import (
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"strconv"
	"strings"
)

type Fasilitas struct {
	IDFasilitas   int    `json:"id_fasilitas"`
	NamaFasilitas string `json:"nama_fasilitas"`
	Kondisi       string `json:"kondisi"`
	Jumlah        int    `json:"jumlah"`
}

type updateFasilitasRequest struct {
	NamaFasilitas *string `json:"nama_fasilitas"`
	Kondisi       *string `json:"kondisi"`
	Jumlah        *int    `json:"jumlah"`
}

func GetFasilitasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	rows, err := config.DB.Query("SELECT id_fasilitas, nama_fasilitas, COALESCE(kondisi, ''), COALESCE(jumlah, 1) FROM fasilitas")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listFasilitas []Fasilitas
	for rows.Next() {
		var f Fasilitas
		if err := rows.Scan(&f.IDFasilitas, &f.NamaFasilitas, &f.Kondisi, &f.Jumlah); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		listFasilitas = append(listFasilitas, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listFasilitas)
}

func CreateFasilitasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var f Fasilitas
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	f.NamaFasilitas = strings.TrimSpace(f.NamaFasilitas)
	if f.NamaFasilitas == "" {
		http.Error(w, `{"message":"Nama fasilitas wajib diisi"}`, http.StatusBadRequest)
		return
	}
	if f.Jumlah < 0 {
		http.Error(w, `{"message":"Jumlah fasilitas tidak boleh negatif"}`, http.StatusBadRequest)
		return
	}
	if f.Jumlah == 0 {
		f.Jumlah = 1
	}

	if strings.TrimSpace(f.Kondisi) == "" {
		f.Kondisi = "Baik"
	}

	query := "INSERT INTO fasilitas (nama_fasilitas, kondisi, jumlah) VALUES (?, ?, ?)"
	_, err = config.DB.Exec(query, f.NamaFasilitas, f.Kondisi, f.Jumlah)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Fasilitas berhasil ditambahkan!",
	})
}

func UpdateFasilitasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idFasilitas, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idFasilitas <= 0 {
		http.Error(w, `{"message":"ID fasilitas tidak valid"}`, http.StatusBadRequest)
		return
	}

	var req updateFasilitasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	bagianSet := []string{}
	argumen := []interface{}{}

	if req.NamaFasilitas != nil {
		nama := strings.TrimSpace(*req.NamaFasilitas)
		if nama == "" {
			http.Error(w, `{"message":"Nama fasilitas tidak boleh kosong"}`, http.StatusBadRequest)
			return
		}
		bagianSet = append(bagianSet, "nama_fasilitas = ?")
		argumen = append(argumen, nama)
	}
	if req.Kondisi != nil {
		kondisi := strings.TrimSpace(*req.Kondisi)
		if kondisi == "" {
			http.Error(w, `{"message":"Kondisi fasilitas tidak boleh kosong"}`, http.StatusBadRequest)
			return
		}
		bagianSet = append(bagianSet, "kondisi = ?")
		argumen = append(argumen, kondisi)
	}
	if req.Jumlah != nil {
		if *req.Jumlah < 0 {
			http.Error(w, `{"message":"Jumlah fasilitas tidak boleh negatif"}`, http.StatusBadRequest)
			return
		}
		bagianSet = append(bagianSet, "jumlah = ?")
		argumen = append(argumen, *req.Jumlah)
	}

	if len(bagianSet) == 0 {
		http.Error(w, `{"message":"Tidak ada data yang diubah"}`, http.StatusBadRequest)
		return
	}

	var ada int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM fasilitas WHERE id_fasilitas = ?", idFasilitas).Scan(&ada)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa fasilitas: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if ada == 0 {
		http.Error(w, `{"message":"Fasilitas tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	query := "UPDATE fasilitas SET " + strings.Join(bagianSet, ", ") + " WHERE id_fasilitas = ?"
	argumen = append(argumen, idFasilitas)
	_, err = config.DB.Exec(query, argumen...)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengubah fasilitas: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Fasilitas berhasil diubah!",
	})
}

func DeleteFasilitasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	idFasilitas, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idFasilitas <= 0 {
		http.Error(w, `{"message":"ID fasilitas tidak valid"}`, http.StatusBadRequest)
		return
	}

	var ada int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM fasilitas WHERE id_fasilitas = ?", idFasilitas).Scan(&ada)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa fasilitas: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if ada == 0 {
		http.Error(w, `{"message":"Fasilitas tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("DELETE FROM fasilitas WHERE id_fasilitas = ?", idFasilitas)
	if err != nil {
		http.Error(w, `{"message":"Gagal menghapus fasilitas: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Fasilitas berhasil dihapus!",
	})
}

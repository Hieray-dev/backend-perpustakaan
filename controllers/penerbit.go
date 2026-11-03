package controllers

import (
	"encoding/json"
	"net/http"
	"perpustakaan-api/config"
	"strconv"
	"strings"
)

type Penerbit struct {
	IDPenerbit   int    `json:"id_penerbit"`
	NamaPenerbit string `json:"nama_penerbit"`
}

func GetPenerbitHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id_penerbit, nama_penerbit FROM penerbit")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listPenerbit []Penerbit
	for rows.Next() {
		var p Penerbit
		if err := rows.Scan(&p.IDPenerbit, &p.NamaPenerbit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		listPenerbit = append(listPenerbit, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listPenerbit)
}

func CreatePenerbitHandler(w http.ResponseWriter, r *http.Request) {
	var p Penerbit
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	p.NamaPenerbit = strings.TrimSpace(p.NamaPenerbit)
	if p.NamaPenerbit == "" {
		http.Error(w, `{"message":"Nama penerbit wajib diisi"}`, http.StatusBadRequest)
		return
	}

	query := "INSERT INTO penerbit (nama_penerbit) VALUES (?)"
	_, err = config.DB.Exec(query, p.NamaPenerbit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Penerbit berhasil ditambahkan!",
	})
}

func UpdatePenerbitHandler(w http.ResponseWriter, r *http.Request) {
	idPenerbit, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idPenerbit <= 0 {
		http.Error(w, `{"message":"ID penerbit tidak valid"}`, http.StatusBadRequest)
		return
	}

	var p Penerbit
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	p.NamaPenerbit = strings.TrimSpace(p.NamaPenerbit)
	if p.NamaPenerbit == "" {
		http.Error(w, `{"message":"Nama penerbit tidak boleh kosong"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM penerbit WHERE id_penerbit = ?", idPenerbit).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa penerbit: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Penerbit tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("UPDATE penerbit SET nama_penerbit = ? WHERE id_penerbit = ?", p.NamaPenerbit, idPenerbit)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengubah penerbit: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Penerbit berhasil diubah!",
	})
}

func DeletePenerbitHandler(w http.ResponseWriter, r *http.Request) {
	idPenerbit, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || idPenerbit <= 0 {
		http.Error(w, `{"message":"ID penerbit tidak valid"}`, http.StatusBadRequest)
		return
	}

	var jumlah int
	err = config.DB.QueryRow("SELECT COUNT(*) FROM penerbit WHERE id_penerbit = ?", idPenerbit).Scan(&jumlah)
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa penerbit: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if jumlah == 0 {
		http.Error(w, `{"message":"Penerbit tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	_, err = config.DB.Exec("DELETE FROM penerbit WHERE id_penerbit = ?", idPenerbit)
	if err != nil {
		http.Error(w, `{"message":"Gagal menghapus penerbit: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Penerbit berhasil dihapus!",
	})
}

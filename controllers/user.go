package controllers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/models"
)

func CreatePetugas(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengenkripsi password"}`, http.StatusInternalServerError)
		return
	}

	idRolePetugas := 2
	if req.IDRole != 0 {
		idRolePetugas = req.IDRole
	}

	query := `INSERT INTO user (username, password, id_role, id_shift, is_active) VALUES (?, ?, ?, ?, 1)`
	_, err = config.DB.Exec(query, req.Username, string(hashedPassword), idRolePetugas, req.IDShift)
	if err != nil {
		http.Error(w, `{"message":"Gagal menambah petugas: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Petugas berhasil ditambahkan",
	})
}

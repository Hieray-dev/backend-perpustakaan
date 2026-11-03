package controllers

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/policy"
)

type UpdateRoleRequest struct {
	IDUser int `json:"id_user"`
	IDRole int `json:"id_role"`
}

type ResetPasswordRequest struct {
	IDUser      int    `json:"id_user"`
	Username    string `json:"username"`
	NewPassword string `json:"new_password"`
}

func UpdateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRoleRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	if req.IDRole != policy.RoleAdmin && req.IDRole != policy.RolePetugas && req.IDRole != policy.RoleAnggota {
		http.Error(w, `{"message":"Role tidak valid (harus 1=Admin, 2=Petugas, atau 3=Anggota)"}`, http.StatusBadRequest)
		return
	}

	query := "UPDATE user SET id_role = ? WHERE id_user = ?"
	result, err := config.DB.Exec(query, req.IDRole, req.IDUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "User tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role user berhasil diperbarui!",
	})
}

func AdminResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.NewPassword == "" || (req.IDUser <= 0 && req.Username == "") {
		http.Error(w, `{"message":"Format JSON tidak valid atau data kosong"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"message":"Gagal memproses password"}`, http.StatusInternalServerError)
		return
	}

	var query string
	var identifier interface{}
	if req.IDUser > 0 {
		query = "UPDATE user SET password = ? WHERE id_user = ?"
		identifier = req.IDUser
	} else {
		query = "UPDATE user SET password = ? WHERE username = ?"
		identifier = req.Username
	}

	result, err := config.DB.Exec(query, string(hashedPassword), identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, `{"message":"User tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password user berhasil di-reset oleh Admin!",
	})
}

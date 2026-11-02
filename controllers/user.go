package controllers

import (
	"database/sql"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/models"
	"perpustakaan-api/policy"
)

func GetAnggotaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := config.DB.Query(`
		SELECT u.id_user, u.username,
		       COALESCE(u.id_role, 0), COALESCE(r.nama_role, ''),
		       COALESCE(u.is_active, 0)
		FROM user u
		LEFT JOIN role r ON u.id_role = r.id_role
		WHERE u.id_role = ?
		ORDER BY u.id_user`, policy.RoleAnggota)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengambil data anggota: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listAnggota []models.Anggota
	for rows.Next() {
		var a models.Anggota
		err := rows.Scan(&a.IDUser, &a.Username, &a.IDRole, &a.NamaRole, &a.IsActive)
		if err != nil {
			http.Error(w, `{"message":"Gagal membaca data anggota"}`, http.StatusInternalServerError)
			return
		}
		listAnggota = append(listAnggota, a)
	}

	json.NewEncoder(w).Encode(listAnggota)
}

func UpdateStatusAnggotaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdateStatusAnggotaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"message":"Format JSON tidak valid"}`, http.StatusBadRequest)
		return
	}

	if req.IDUser <= 0 {
		http.Error(w, `{"message":"ID user tidak valid"}`, http.StatusBadRequest)
		return
	}
	if req.IsActive == nil {
		http.Error(w, `{"message":"Field is_active wajib diisi (0 atau 1)"}`, http.StatusBadRequest)
		return
	}
	if *req.IsActive != 0 && *req.IsActive != 1 {
		http.Error(w, `{"message":"Nilai is_active harus 0 atau 1"}`, http.StatusBadRequest)
		return
	}

	var idRole int
	err = config.DB.QueryRow("SELECT COALESCE(id_role, 0) FROM user WHERE id_user = ?", req.IDUser).Scan(&idRole)
	if err == sql.ErrNoRows {
		http.Error(w, `{"message":"Anggota tidak ditemukan"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"message":"Gagal memeriksa anggota: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if idRole != policy.RoleAnggota {
		http.Error(w, `{"message":"User tersebut bukan anggota"}`, http.StatusBadRequest)
		return
	}

	_, err = config.DB.Exec("UPDATE user SET is_active = ? WHERE id_user = ? AND id_role = ?", *req.IsActive, req.IDUser, policy.RoleAnggota)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengubah status anggota: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Status anggota berhasil diperbarui!",
	})
}

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

package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"perpustakaan-api/config"
	"perpustakaan-api/models"
	"perpustakaan-api/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// /register publik: id_role tidak diterima dari input, akun selalu dibuat sebagai anggota (3)
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IDShift  *int   `json:"id_shift"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Gagal memproses enkripsi password", http.StatusInternalServerError)
		return
	}

	query := "INSERT INTO user (username, password, id_role, id_shift, is_active) VALUES (?, ?, 3, ?, 1)"
	_, err = config.DB.Exec(query, req.Username, string(hashedPassword), req.IDShift)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Register berhasil!",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req models.User
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	var dbUser models.User
	var namaShift, jamMasuk, jamKeluar string

	query := `
		SELECT 
			u.id_user, u.username, u.password, u.id_role, u.is_active,
			COALESCE(js.nama_shift, 'Tidak Ada Shift'),
			COALESCE(js.jam_masuk, '00:00:00'),
			COALESCE(js.jam_keluar, '00:00:00')
		FROM user u
		LEFT JOIN jadwal_shift js ON u.id_shift = js.id_shift
		WHERE u.username = ?`

	err = config.DB.QueryRow(query, req.Username).Scan(
		&dbUser.IDUser, &dbUser.Username, &dbUser.Password, &dbUser.IDRole, &dbUser.IsActive,
		&namaShift, &jamMasuk, &jamKeluar,
	)
	if err != nil {
		http.Error(w, "Username tidak ditemukan", http.StatusUnauthorized)
		return
	}

	if dbUser.IsActive == 0 {
		http.Error(w, "Akun anda telah dinonaktifkan oleh admin!", http.StatusForbidden)
		return
	}

	errBcrypt := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password))
	if errBcrypt != nil {
		http.Error(w, "Password salah!", http.StatusUnauthorized)
		return
	}

	waktuSekarang := time.Now()
	jamLoginStr := waktuSekarang.Format("15:04:05")
	tanggalHariIni := waktuSekarang.Format("2006-01-02")

	statusKehadiran := "Tidak Ada Absensi"

	if namaShift != "Tidak Ada Shift" {
		tLogin, _ := time.Parse("15:04:05", jamLoginStr)
		tMasuk, _ := time.Parse("15:04:05", jamMasuk)

		detikLogin := tLogin.Hour()*3600 + tLogin.Minute()*60 + tLogin.Second()
		detikMasuk := tMasuk.Hour()*3600 + tMasuk.Minute()*60 + tMasuk.Second()

		if detikLogin > detikMasuk {
			statusKehadiran = "Terlambat"
		} else {
			statusKehadiran = "Tepat Waktu"
		}

		queryAbsensi := "INSERT INTO absensi (id_user, tanggal, status_kehadiran) VALUES (?, ?, ?)"
		_, errAbsensi := config.DB.Exec(queryAbsensi, dbUser.IDUser, tanggalHariIni, statusKehadiran)
		if errAbsensi != nil {
			log.Println("Gagal menyimpan absensi:", errAbsensi)
		}
	}

	token, err := utils.GenerateToken(dbUser.IDUser, dbUser.Username, dbUser.IDRole)
	if err != nil {
		http.Error(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":          "Login berhasil",
		"token":            token,
		"id_user":          dbUser.IDUser,
		"username":         dbUser.Username,
		"id_role":          dbUser.IDRole,
		"nama_shift":       namaShift,
		"jam_masuk_shift":  jamMasuk,
		"jam_login":        jamLoginStr,
		"status_kehadiran": statusKehadiran,
	})
}

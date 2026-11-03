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
	var idShift int
	var namaRole, namaShift, jamMasuk, jamKeluar string

	query := `
		SELECT 
			u.id_user, u.username, u.password, u.id_role, COALESCE(r.nama_role, ''), u.is_active,
			COALESCE(u.id_shift, 0),
			COALESCE(js.nama_shift, 'Tidak Ada Shift'),
			COALESCE(js.jam_masuk, '00:00:00'),
			COALESCE(js.jam_keluar, '00:00:00')
		FROM user u
		LEFT JOIN role r ON u.id_role = r.id_role
		LEFT JOIN jadwal_shift js ON u.id_shift = js.id_shift
		WHERE u.username = ?`

	err = config.DB.QueryRow(query, req.Username).Scan(
		&dbUser.IDUser, &dbUser.Username, &dbUser.Password, &dbUser.IDRole, &namaRole, &dbUser.IsActive,
		&idShift, &namaShift, &jamMasuk, &jamKeluar,
	)
	if err != nil {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}

	errBcrypt := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(req.Password))
	if errBcrypt != nil {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}

	if dbUser.IsActive == 0 {
		http.Error(w, "Akun anda telah dinonaktifkan oleh admin!", http.StatusForbidden)
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

		var statusTercatat string
		errCek := config.DB.QueryRow(
			"SELECT status_kehadiran FROM absensi WHERE id_user = ? AND tanggal = ? LIMIT 1",
			dbUser.IDUser, tanggalHariIni,
		).Scan(&statusTercatat)

		if errCek == nil {
			statusKehadiran = statusTercatat
		} else {
			queryAbsensi := "INSERT INTO absensi (id_user, id_shift, tanggal, status_kehadiran) VALUES (?, ?, ?, ?)"
			_, errAbsensi := config.DB.Exec(queryAbsensi, dbUser.IDUser, idShift, tanggalHariIni, statusKehadiran)
			if errAbsensi != nil {
				log.Println("Gagal menyimpan absensi:", errAbsensi)
			}
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
		"nama_role":        namaRole,
		"nama_shift":       namaShift,
		"jam_masuk_shift":  jamMasuk,
		"jam_login":        jamLoginStr,
		"status_kehadiran": statusKehadiran,
	})
}

func GetAbsensiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, `{"message":"Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rows, err := config.DB.Query(`
		SELECT a.id_absensi, a.id_user, u.username,
		       COALESCE(a.id_shift, u.id_shift, 0), COALESCE(js.nama_shift, ''),
		       a.tanggal, COALESCE(js.jam_masuk, ''), a.jam_masuk_aktual,
		       COALESCE(a.jam_keluar_aktual, ''), COALESCE(a.status_kehadiran, '')
		FROM absensi a
		JOIN user u ON a.id_user = u.id_user
		LEFT JOIN jadwal_shift js ON COALESCE(a.id_shift, u.id_shift) = js.id_shift
		ORDER BY a.id_absensi DESC`)
	if err != nil {
		http.Error(w, `{"message":"Gagal mengambil data absensi: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listAbsensi []models.Absensi
	for rows.Next() {
		var a models.Absensi
		err := rows.Scan(
			&a.IDAbsensi, &a.IDUser, &a.Username,
			&a.IDShift, &a.NamaShift,
			&a.Tanggal, &a.JamMasukShift, &a.JamMasukAktual,
			&a.JamKeluarAktual, &a.StatusKehadiran,
		)
		if err != nil {
			http.Error(w, `{"message":"Gagal membaca data absensi"}`, http.StatusInternalServerError)
			return
		}
		listAbsensi = append(listAbsensi, a)
	}

	json.NewEncoder(w).Encode(listAbsensi)
}

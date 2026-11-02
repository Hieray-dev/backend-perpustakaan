package models

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type Absensi struct {
	IDAbsensi       int    `json:"id_absensi"`
	IDUser          int    `json:"id_user"`
	Username        string `json:"username"`
	IDShift         int    `json:"id_shift"`
	NamaShift       string `json:"nama_shift"`
	Tanggal         string `json:"tanggal"`
	JamMasukShift   string `json:"jam_masuk_shift"`
	JamMasukAktual  string `json:"jam_masuk_aktual"`
	JamKeluarAktual string `json:"jam_keluar_aktual"`
	StatusKehadiran string `json:"status_kehadiran"`
}

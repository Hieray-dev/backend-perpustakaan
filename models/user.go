package models

type User struct {
	IDUser   int    `json:"id_user"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	IDRole   int    `json:"id_role"`
	IDShift  *int   `json:"id_shift,omitempty"`
	IsActive int    `json:"is_active"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IDRole   int    `json:"id_role"`
	IDShift  int    `json:"id_shift"`
}

type LoginShiftResponse struct {
	Message         string `json:"message"`
	Username        string `json:"username"`
	NamaShift       string `json:"nama_shift"`
	JamMasukShift   string `json:"jam_masuk_shift"`
	JamLogin        string `json:"jam_login"`
	StatusKehadiran string `json:"status_kehadiran"`
}

type Anggota struct {
	IDUser   int    `json:"id_user"`
	Username string `json:"username"`
	IDRole   int    `json:"id_role"`
	NamaRole string `json:"nama_role"`
	IsActive int    `json:"is_active"`
}

type UpdateStatusAnggotaRequest struct {
	IDUser   int  `json:"id_user"`
	IsActive *int `json:"is_active"`
}

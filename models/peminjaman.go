package models

type Riwayat struct {
	IDPeminjaman  int    `json:"id_peminjaman"`
	KodeTransaksi string `json:"kode_transaksi"`
	JudulBuku     string `json:"judul_buku"`
	IDPeminjam    int    `json:"id_peminjam"`
	Peminjam      string `json:"peminjam"`
	TanggalPinjam string `json:"tanggal_peminjaman"`
	Status        string `json:"status"`
	Pustakawan    string `json:"pustakawan"`
	StatusPetugas string `json:"status_petugas"`
}

type BookingRequest struct {
	IDBuku int `json:"id_buku"`
}

type ConfirmPinjamRequest struct {
	IDPeminjaman int `json:"id_peminjaman"`
}

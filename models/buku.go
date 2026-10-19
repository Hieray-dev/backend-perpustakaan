package models

type Buku struct {
	IDBuku      int    `json:"id_buku"`
	Judul       string `json:"judul"`
	Penulis     string `json:"penulis"`
	TahunTerbit int    `json:"tahun_terbit"`
	Deskripsi   string `json:"deskripsi"`
	Gambar      string `json:"gambar"`
	Stok        int    `json:"stok"`
	IDPenerbit  int    `json:"id_penerbit"`
	Penerbit    string `json:"penerbit"`
	IDGenre     int    `json:"id_genre"`
	Genre       string `json:"genre"`
}

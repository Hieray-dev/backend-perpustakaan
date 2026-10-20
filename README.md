# perpustakaan-api

REST API backend untuk sistem perpustakaan. Dibuat dengan Go dan `net/http` standar tanpa
framework, dengan MySQL/MariaDB sebagai basis data dan JWT untuk autentikasi.

Antarmuka frontend dikembangkan terpisah; repo ini hanya menyediakan API.

## Daftar isi

- [Fitur](#fitur)
- [Teknologi](#teknologi)
- [Struktur proyek](#struktur-proyek)
- [Persiapan](#persiapan)
- [Instalasi](#instalasi)
- [Konfigurasi `.env`](#konfigurasi-env)
- [Menjalankan](#menjalankan)
- [Autentikasi](#autentikasi)
- [Daftar endpoint](#daftar-endpoint)
- [Contoh pemakaian](#contoh-pemakaian)
- [Peran pengguna](#peran-pengguna)
- [Alur transaksi](#alur-transaksi)
- [Perubahan terbaru](#perubahan-terbaru)
- [Rencana pengembangan](#rencana-pengembangan)

## Fitur

**Autentikasi dan kehadiran**

- Login dengan JWT (masa berlaku 24 jam), password di-hash dengan bcrypt
- Pencatatan absensi saat petugas login direncanakan dengan status `Tepat Waktu` atau
  `Terlambat` dihitung dari jam masuk shift — **belum aktif**, query absensi masih perlu
  diperbaiki (lihat [Rencana pengembangan](#rencana-pengembangan))
- Akun nonaktif ditolak saat login

**Katalog (akses publik)**

- Daftar buku beserta stok
- Daftar genre, penerbit, dan fasilitas

**Transaksi peminjaman**

- Anggota membuat *booking* buku dengan kode transaksi; batas pengambilan 24 jam baru
  dihitung, belum ditegakkan otomatis (kolom `batas_waktu` masih bertipe tanggal)
- Petugas mengonfirmasi booking menjadi peminjaman, jatuh tempo 7 hari
- Pengembalian buku dengan perhitungan denda otomatis (Rp2.000 per hari terlambat) dan
  pengembalian stok
- Riwayat seluruh transaksi

**Manajemen oleh admin**

- Menambah akun petugas beserta shiftnya
- Mengubah role pengguna
- Reset password pengguna

## Teknologi

| Bagian | Dipakai |
|---|---|
| Bahasa | Go 1.26 |
| HTTP | `net/http` (standard library, tanpa framework) |
| Database | MySQL / MariaDB |
| Driver | `github.com/go-sql-driver/mysql` v1.10.0 |
| JWT | `github.com/golang-jwt/jwt/v5` v5.3.1 |
| Hash password | `golang.org/x/crypto/bcrypt` |
| Konfigurasi | `github.com/joho/godotenv` v1.5.1 |

## Struktur proyek

```
perpustakaan-api/
├── main.go                 # entry point, CORS middleware, HTTP server
├── config/
│   └── db.go               # koneksi MySQL, membaca .env
├── controllers/            # handler HTTP per domain
│   ├── auth.go             # register, login, absensi
│   ├── buku.go             # katalog dan tambah buku
│   ├── peminjaman.go       # booking, konfirmasi, riwayat
│   ├── pengembalian.go     # pengembalian dan denda
│   ├── admin.go            # ubah role, reset password
│   ├── user.go             # tambah petugas
│   ├── fasilitas.go
│   ├── genre.go
│   └── penerbit.go
├── middleware/
│   ├── auth.go             # verifikasi JWT, menaruh claims di context
│   └── role.go             # AdminOnly, PetugasOrAdmin
├── models/                 # struct request dan response
├── policy/
│   └── policy.go           # konstanta role
├── routes/
│   └── routes.go           # pendaftaran seluruh rute
├── utils/
│   └── jwt.go              # generate dan validasi token
├── uploads/                # file statis, disajikan di /uploads/
└── schema.sql              # struktur tabel database
```

### Skema database

11 tabel: `user`, `role`, `jadwal_shift`, `absensi`, `buku`, `genre`, `penerbit`, `status`,
`peminjaman`, `pengembalian`, `fasilitas`.

Relasi utama:

```
role ──< user >── jadwal_shift
           │
           ├──< absensi
           │
           └──< peminjaman >── buku >── genre
                                  ├── penerbit
                                  └── status
```

## Persiapan

- Go 1.26 atau lebih baru
- MySQL 8 atau MariaDB 10.4 atau lebih baru
- Git

## Instalasi

**1. Clone repositori**

```bash
git clone git@github.com:Hieray-dev/backend-perpustakaan.git
cd backend-perpustakaan
```

**2. Unduh dependensi**

```bash
go mod download
```

**3. Buat database**

```bash
mysql -u root -p -e "CREATE DATABASE db_perpustakaan;"
```

**4. Import struktur tabel + data dasar**

> **Peringatan:** `schema.sql` dimulai dengan `DROP TABLE IF EXISTS`. Jangan jalankan pada
> database yang sudah berisi data — tabel lama akan terhapus.

```bash
mysql -u root -p db_perpustakaan < schema.sql
```

`schema.sql` hanya berisi struktur tabel tanpa data. Tanpa data dasar, `POST /register`
gagal karena foreign key `user.id_role` → `role.id_role`. Isi minimalnya:

```sql
INSERT INTO role (id_role, nama_role) VALUES
  (1, 'Admin'), (2, 'Petugas'), (3, 'Anggota');

INSERT INTO jadwal_shift (id_shift, nama_shift, jam_masuk, jam_keluar) VALUES
  (1, 'Shift Pagi', '08:00:00', '15:00:00'),
  (2, 'Shift Siang', '15:00:00', '21:00:00');
```

Kalau database sudah berjalan dan kamu ingin memperbarui berkas strukturnya:

```bash
mysqldump -u root -p --no-data --skip-comments db_perpustakaan > schema.sql
```

**5. Siapkan file `.env`**

```bash
cp .env.example .env
```

Lalu isi nilainya seperti dijelaskan di bawah, dan jalankan server (lihat bagian
[Menjalankan](#menjalankan)).

**6. Buat akun admin pertama**

Tidak ada endpoint publik untuk membuat admin. Setelah server berjalan, daftarkan akun
biasa lewat `POST /register`, lalu ubah rolenya lewat database:

```sql
UPDATE user SET id_role = 1 WHERE username = 'nama_akun';
```

## Konfigurasi `.env`

```env
DB_USER=root
DB_PASS=
DB_HOST=localhost
DB_PORT=3306
DB_NAME=db_perpustakaan

JWT_SECRET=
SERVER_PORT=:8080
```

| Variabel | Keterangan |
|---|---|
| `DB_USER` | Pengguna MySQL |
| `DB_PASS` | Password MySQL |
| `DB_HOST` | Host MySQL |
| `DB_PORT` | Port MySQL, umumnya `3306` |
| `DB_NAME` | Nama database, gunakan `db_perpustakaan` |
| `JWT_SECRET` | Kunci penandatanganan token. **Wajib diisi**, dan jangan diganti setelah ada token aktif karena semua token lama langsung tidak berlaku |
| `SERVER_PORT` | Port server. Boleh `8080` atau `:8080`, keduanya diterima |

File `.env` sudah masuk `.gitignore` dan tidak boleh di-commit.

## Menjalankan

```bash
go run .
```

Kalau berhasil akan muncul:

```
Koneksi Database Berhasil!
Server running on http://0.0.0.0:8080...
```

Server mengikat ke `0.0.0.0`, jadi bisa diakses dari perangkat lain di jaringan yang sama
lewat alamat IP komputermu.

Jalankan dari **root repo**, karena `.env` dan folder `uploads/` dibaca lewat path relatif.
CORS masih terbuka untuk semua origin (`*`) — cocok untuk pengembangan, sebaiknya dibatasi
saat produksi.

Build ke satu file binary:

```bash
go build -o perpustakaan-api .
./perpustakaan-api
```

## Autentikasi

Alur standar:

1. Kirim `POST /login` dengan `username` dan `password`
2. Server membalas `token` berupa JWT
3. Sertakan token di setiap request yang butuh login:

```
Authorization: Bearer <token>
```

Isi token memuat `id_user`, `username`, dan `id_role`. Masa berlakunya 24 jam.

Ada empat tingkat akses pada endpoint:

| Tingkat | Arti |
|---|---|
| Publik | Tidak perlu token |
| Login | Perlu token, semua role boleh |
| Petugas/Admin | Perlu token dengan `id_role` 1 atau 2 |
| Admin | Perlu token dengan `id_role` 1 |

## Daftar endpoint

> Catatan: endpoint daftar mengembalikan `null` (bukan `[]`) bila tabelnya kosong, karena
> slice hasil query belum diinisialisasi.

### Publik

| Method | Endpoint | Keterangan |
|---|---|---|
| `POST` | `/register` | Mendaftar akun baru (selalu berperan anggota); `id_shift` opsional |
| `POST` | `/login` | Login, mengembalikan token |
| `GET` | `/buku` | Daftar buku dan stok |
| `GET` | `/genre` | Daftar genre |
| `GET` | `/penerbit` | Daftar penerbit |
| `GET` | `/fasilitas` | Daftar fasilitas |
| `GET` | `/uploads/<namafile>` | File statis, misalnya gambar buku |

### Perlu login

| Method | Endpoint | Keterangan |
|---|---|---|
| `POST` | `/booking` | Membuat booking buku |
| `GET` | `/histori/riwayat` | Riwayat peminjaman beserta `status_petugas` |

### Petugas atau Admin

| Method | Endpoint | Keterangan |
|---|---|---|
| `POST` | `/buku/create` | Menambah buku |
| `POST` | `/buku/upload` | Mengunggah gambar buku (multipart, maks 2 MB) |
| `PUT` | `/buku/{id}` | Mengubah sebagian field buku |
| `DELETE` | `/buku/{id}` | Menghapus buku (ditolak bila masih dibooking/dipinjam) |
| `POST` | `/genre/create` | Menambah genre |
| `POST` | `/penerbit/create` | Menambah penerbit |
| `POST` | `/peminjaman/confirm` | Mengonfirmasi booking menjadi peminjaman |
| `POST` | `/pengembalian` | Mencatat pengembalian dan menghitung denda |

### Admin

| Method | Endpoint | Keterangan |
|---|---|---|
| `POST` | `/petugas/create` | Membuat akun dengan role dari body (1/2/3) beserta shift |
| `POST` | `/fasilitas/create` | Menambah fasilitas |
| `PUT` | `/admin/user/role` | Mengubah role pengguna |
| `PUT` | `/admin/user/reset-password` | Reset password pengguna |

## Contoh pemakaian

**Login**

```bash
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"rahasia"}'
```

```json
{
  "message": "Login berhasil",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "id_user": 1,
  "username": "admin",
  "id_role": 1,
  "nama_shift": "Pagi",
  "jam_masuk_shift": "08:00:00",
  "jam_login": "07:52:10",
  "status_kehadiran": "Tepat Waktu"
}
```

**Lihat katalog**

```bash
curl http://localhost:8080/buku
```

```json
[
  {
    "id_buku": 1,
    "judul": "Belajar MySQL & MariaDB untuk Pemula",
    "penulis": "Dian Sastro",
    "tahun_terbit": 0,
    "deskripsi": "Panduan lengkap dasar-dasar MySQL dan MariaDB untuk pemula.",
    "gambar": "mysql.jpg",
    "stok": 1,
    "id_penerbit": 1,
    "penerbit": "Informatika Bandung",
    "id_genre": 1,
    "genre": "Teknologi & Koding"
  }
]
```

`penerbit` dan `genre` berisi **nama** hasil penggabungan dengan tabel `penerbit` dan
`genre`. Pasangan angkanya ada di `id_penerbit` dan `id_genre` — angka itu yang dipakai
saat menambah buku (endpoint ubah buku belum tersedia). `tahun_terbit` bernilai `0` selama
belum diisi.

**Tambah buku** (perlu token petugas atau admin)

```bash
curl -X POST http://localhost:8080/buku/create \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "judul": "Algoritma dan Pemrograman",
    "penulis": "Rinaldi Munir",
    "tahun_terbit": 2011,
    "deskripsi": "Buku algoritma dan struktur data.",
    "gambar": "algoritma.jpg",
    "stok": 3,
    "id_penerbit": 1,
    "id_genre": 1
  }'
```

`tahun_terbit`, `id_penerbit`, dan `id_genre` boleh dikosongkan. Daftar angka yang sah bisa
diambil dari `GET /penerbit` dan `GET /genre`.

**Membuat booking**

```bash
curl -X POST http://localhost:8080/booking \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id_user":2,"id_buku":1}'
```

```json
{
  "message": "Booking berhasil",
  "kode_transaksi": "BK-1810063500",
  "batas_waktu_ambil": "2026-10-19 10:30:00"
}
```

**Konfirmasi peminjaman** (perlu token petugas atau admin)

```bash
curl -X POST http://localhost:8080/peminjaman/confirm \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id_peminjaman":1,"id_petugas":2}'
```

**Pengembalian** (perlu token petugas atau admin)

```bash
curl -X POST http://localhost:8080/pengembalian \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id_peminjaman":1,"id_petugas":2}'
```

```json
{
  "message": "Buku berhasil dikembalikan!",
  "id_peminjaman": 1,
  "tanggal_pengembalian": "2026-10-25",
  "denda": 0,
  "status_transaksi": "dikembalikan"
}
```

## Peran pengguna

| `id_role` | Peran | Hak akses |
|---|---|---|
| 1 | Admin | Seluruh endpoint, termasuk kelola pengguna |
| 2 | Petugas | Katalog (baca dan tambah), konfirmasi pinjam, pengembalian |
| 3 | Anggota | Katalog, booking, dan riwayat (saat ini masih seluruh transaksi; target: hanya milik sendiri) |

Nilai role didefinisikan di `policy/policy.go`.

## Alur transaksi

```
Anggota booking  →  Petugas konfirmasi  →  Petugas catat pengembalian
      ↓                     ↓                          ↓
 stok berkurang      jatuh tempo 7 hari         stok bertambah
 batas ambil 24 jam  status: dipinjam           denda dihitung
 status: booking
```

Denda dihitung per hari keterlambatan, diukur dari **batas kembali** (7 hari sejak buku
diserahkan, tersimpan di kolom `tanggal_kembali`). Tarif per hari dan perhitungannya ada di
`controllers/pengembalian.go`.

## Perubahan terbaru

**20 Oktober 2026**

- Pencatatan absensi saat login diperbaiki: kolom `status` diganti `status_kehadiran`, dan
  kegagalan menyimpan kini dicatat di log (sebelumnya gagal secara senyap)
- `PUT /buku/{id}` mengubah sebagian field buku; `DELETE /buku/{id}` menghapus buku, ditolak
  bila buku masih dibooking atau dipinjam
- `POST /buku/upload` menerima gambar (maks 2 MB, jpg/png/webp), memeriksa isi berkas, lalu
  menyimpannya dengan nama acak di `uploads/` dan mengembalikan URL
- Validasi tambah/ubah buku: judul wajib, stok tidak negatif, dan `id_penerbit`/`id_genre`/
  `id_status` harus ada bila diisi; kesalahan input dibalas `400`, bukan `500`
- `uploads/firefly.png` (gambar contoh lama) dihapus dari repo

**19 Oktober 2026**

- Login hanya memakai pemeriksaan `bcrypt`; perbandingan password sebagai teks biasa dihapus
- `POST /register` selalu membuat akun berperan anggota — `id_role` tidak lagi diterima dari
  permintaan
- `JWT_SECRET` wajib diisi: server menolak dijalankan kalau kosong (nilai cadangan di kode
  sudah dihapus)
- Denda pengembalian dihitung dari batas kembali (`tanggal_kembali`), bukan batas pengambilan
  (`batas_waktu`)
- `GET /buku` mengirim nama penerbit dan genre hasil penggabungan tabel, ditambah
  `id_penerbit`, `id_genre`, dan `tahun_terbit`; `POST /buku/create` menyimpan kolom itu
- Kolom `tahun_terbit` ditambahkan pada tabel `buku` (lihat `schema.sql`)

Kalau database sudah ada dari sebelumnya, tambahkan kolom barunya dengan:

```sql
ALTER TABLE buku ADD COLUMN tahun_terbit int(11) DEFAULT NULL AFTER penulis;
```

## Rencana pengembangan

- Endpoint ubah dan hapus untuk genre, penerbit, dan fasilitas
- Pagination dan pencarian pada katalog
- Endpoint daftar absensi dan daftar pengembalian
- Pengelolaan data anggota
- Pembatalan otomatis booking yang melewati batas pengambilan
- Pengujian otomatis

## Kontributor

- [Hieray-dev](https://github.com/Hieray-dev)

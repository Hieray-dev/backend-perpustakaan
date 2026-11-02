package routes

import (
	"net/http"
	"perpustakaan-api/controllers"
	"perpustakaan-api/middleware"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	mux.HandleFunc("/register", controllers.RegisterHandler)
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/buku", controllers.GetKatalogBuku)
	mux.HandleFunc("/fasilitas", controllers.GetFasilitasHandler)
	mux.HandleFunc("/penerbit", controllers.GetPenerbitHandler)
	mux.HandleFunc("/genre", controllers.GetGenreHandler)

	mux.Handle("/booking", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateBooking)))
	mux.Handle("/histori/riwayat", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetRiwayatHandler)))

	mux.Handle("POST /buku/create", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.CreateBukuHandler))))
	mux.Handle("POST /buku/upload", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UploadGambarBukuHandler))))
	mux.Handle("PUT /buku/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UpdateBukuHandler))))
	mux.Handle("DELETE /buku/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.DeleteBukuHandler))))
	mux.Handle("POST /penerbit/create", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.CreatePenerbitHandler))))
	mux.Handle("PUT /penerbit/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UpdatePenerbitHandler))))
	mux.Handle("DELETE /penerbit/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.DeletePenerbitHandler))))
	mux.Handle("POST /genre/create", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.CreateGenreHandler))))
	mux.Handle("PUT /genre/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UpdateGenreHandler))))
	mux.Handle("DELETE /genre/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.DeleteGenreHandler))))
	mux.Handle("/peminjaman/confirm", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.ConfirmPinjam))))
	mux.Handle("GET /pengembalian", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.GetPengembalianHandler))))
	mux.Handle("POST /pengembalian", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.PengembalianHandler))))
	mux.Handle("GET /absensi", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.GetAbsensiHandler))))
	mux.Handle("GET /anggota", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.GetAnggotaHandler))))
	mux.Handle("PUT /anggota/status", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UpdateStatusAnggotaHandler))))

	mux.Handle("POST /fasilitas/create", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.CreateFasilitasHandler))))
	mux.Handle("PUT /fasilitas/{id}", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.UpdateFasilitasHandler))))
	mux.Handle("DELETE /fasilitas/{id}", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.DeleteFasilitasHandler))))
	mux.Handle("/petugas/create", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.CreatePetugas))))
	mux.Handle("/admin/user/role", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.UpdateUserRoleHandler))))
	mux.Handle("/admin/user/reset-password", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.AdminResetPasswordHandler))))

	return mux
}

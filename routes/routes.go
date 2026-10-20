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
	mux.Handle("PUT /buku/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.UpdateBukuHandler))))
	mux.Handle("DELETE /buku/{id}", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.DeleteBukuHandler))))
	mux.Handle("/penerbit/create", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.CreatePenerbitHandler))))
	mux.Handle("/genre/create", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.CreateGenreHandler))))
	mux.Handle("/peminjaman/confirm", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.ConfirmPinjam))))
	mux.Handle("/pengembalian", middleware.AuthMiddleware(middleware.PetugasOrAdmin(http.HandlerFunc(controllers.PengembalianHandler))))

	mux.Handle("/fasilitas/create", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.CreateFasilitasHandler))))
	mux.Handle("/petugas/create", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.CreatePetugas))))
	mux.Handle("/admin/user/role", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.UpdateUserRoleHandler))))
	mux.Handle("/admin/user/reset-password", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(controllers.AdminResetPasswordHandler))))

	return mux
}

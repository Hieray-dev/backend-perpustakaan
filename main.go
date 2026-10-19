package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"perpustakaan-api/config"
	"perpustakaan-api/routes"
	"strings"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	config.ConnectDB()

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET belum diisi di .env, server dihentikan")
	}

	appRoutes := routes.SetupRoutes()

	handlerWithCORS := corsMiddleware(appRoutes)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	address := "0.0.0.0" + port

	fmt.Printf("Server running on http://%s...\n", address)

	err := http.ListenAndServe(address, handlerWithCORS)
	if err != nil {
		fmt.Println("Gagal menjalankan server:", err)
	}
}

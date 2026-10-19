package middleware

import (
	"net/http"
	"perpustakaan-api/policy"
	"perpustakaan-api/utils"
)

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserCtxKey).(*utils.Claims)
		if !ok || claims.IDRole != policy.RoleAdmin {
			http.Error(w, `{"message":"Akses ditolak: Khusus Admin"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func PetugasOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserCtxKey).(*utils.Claims)
		if !ok || (claims.IDRole != policy.RoleAdmin && claims.IDRole != policy.RolePetugas) {
			http.Error(w, `{"message":"Akses ditolak: Khusus Petugas atau Admin"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

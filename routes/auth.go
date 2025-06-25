package routes

import (
	"github.com/gomisroca/gasthaus-backend/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterAuthRoutes(r *mux.Router, dbpool *pgxpool.Pool) {
	sr := r.PathPrefix("/auth").Subrouter()
	h := &handlers.AuthHandler{DB: dbpool}

	sr.HandleFunc("/login", h.Login).Methods("POST")
	sr.HandleFunc("/refresh-token", h.RefreshToken).Methods("GET")
}

package routes

import (
	"net/http"

	"github.com/gomisroca/gasthaus-backend/handlers"
	"github.com/gomisroca/gasthaus-backend/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSpeisekarteRoutes(r *mux.Router, dbpool *pgxpool.Pool) {
	sr := r.PathPrefix("/speisekarte").Subrouter()
	h := &handlers.SpeisekarteHandler{DB: dbpool}

	sr.HandleFunc("/", h.GetItems).Methods("GET")
	sr.Handle("/", middleware.JWTAuth(http.HandlerFunc(h.AddItem))).Methods("POST")
	sr.HandleFunc("/categories", h.GetCategories).Methods("GET")
	sr.Handle("/{id}", middleware.JWTAuth(http.HandlerFunc(h.UpdateItem))).Methods("PUT")
	sr.Handle("/{id}", middleware.JWTAuth(http.HandlerFunc(h.DeleteItem))).Methods("DELETE")
}

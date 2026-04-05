package routes

import (
	"net/http"

	"github.com/gomisroca/gasthaus-backend/handlers"
	"github.com/gomisroca/gasthaus-backend/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSpeisekarteRoutes(r *mux.Router, dbpool *pgxpool.Pool, jwtSecret string) {
	sr := r.PathPrefix("/speisekarte").Subrouter()
	h := &handlers.SpeisekarteHandler{DB: dbpool}
	auth := middleware.JWTAuth(jwtSecret)

	sr.HandleFunc("/", h.GetItems).Methods("GET")
	sr.Handle("/", auth(http.HandlerFunc(h.AddItem))).Methods("POST")
	sr.HandleFunc("/categories", h.GetCategories).Methods("GET")
	sr.HandleFunc("/{id}", h.GetUniqueItem).Methods("GET")
	sr.Handle("/{id}", auth(http.HandlerFunc(h.UpdateItem))).Methods("PUT")
	sr.Handle("/{id}", auth(http.HandlerFunc(h.DeleteItem))).Methods("DELETE")
}
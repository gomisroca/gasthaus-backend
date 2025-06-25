package routes

import (
	"fmt"
	"net/http"

	"github.com/gomisroca/gasthaus-backend/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSpeisekarteRoutes(r *mux.Router, dbpool *pgxpool.Pool) {
	sr := r.PathPrefix("/speisekarte").Subrouter()
	h := &handlers.SpeisekarteHandler{DB: dbpool}

	sr.HandleFunc("/", h.GetItems).Methods("GET")

	sr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Adding items to the Speisekarte")
	}).Methods("POST")

	sr.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		fmt.Fprintf(w, "Updating item with ID: %s", id)
	}).Methods("PUT")

	sr.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		fmt.Fprintf(w, "Deleting item with ID: %s", id)
	}).Methods("DELETE")
}

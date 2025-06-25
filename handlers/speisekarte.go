package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomisroca/gasthaus-backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
)


type SpeisekarteHandler struct {
	DB *pgxpool.Pool
}

func (h *SpeisekarteHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	var items []models.SpeisekarteItem

	rows, err := h.DB.Query(context.Background(), "SELECT id, name FROM items")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.SpeisekarteItem
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomisroca/gasthaus-backend/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)


type SpeisekarteHandler struct {
	DB *pgxpool.Pool
}

func (h *SpeisekarteHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	var items []models.SpeisekarteItem

	rows, err := h.DB.Query(context.Background(), "SELECT id, name, description, price, categories, tags, image, seasonal FROM speisekarte")
	if err != nil {
		log.Printf("Database query failed: %v", err)
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.SpeisekarteItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Categories,
			&item.Tags,
			&item.Image,
			&item.Seasonal,
		); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}


func (h *SpeisekarteHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var item models.SpeisekarteItem

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Generate a new UUID for the item
	newID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	item.ID = newID.String()

	query := `
		INSERT INTO speisekarte (id, name, description, price, categories, tags, image, seasonal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = h.DB.Exec(
		context.Background(),
		query,
		item.ID,
		item.Name,
		item.Description,
		item.Price,
		item.Categories,
		item.Tags,
		item.Image,
		item.Seasonal,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			http.Error(w, "Item with this name already exists", http.StatusConflict)
			return
		}
		log.Printf("Failed to insert item: %v", err)
		http.Error(w, "Failed to insert item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": item.ID})
}

func (h *SpeisekarteHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}

func (h *SpeisekarteHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
}
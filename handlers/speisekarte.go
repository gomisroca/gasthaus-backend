package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gomisroca/gasthaus-backend/models"
	"github.com/gorilla/mux"
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

	query := `
		INSERT INTO speisekarte (name, description, price, categories, tags, image, seasonal)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := h.DB.Exec(
		context.Background(),
		query,
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

	w.WriteHeader(http.StatusOK)
}

func (h *SpeisekarteHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	var updatedItem models.SpeisekarteItem
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE speisekarte
		SET name = $1,
			description = $2,
			price = $3,
			categories = $4,
			tags = $5,
			image = $6,
			seasonal = $7
		WHERE id = $8
		RETURNING id;
	`

	var returnedID string
	err := h.DB.QueryRow(
		context.Background(),
		query,
		updatedItem.Name,
		updatedItem.Description,
		updatedItem.Price,
		updatedItem.Categories,
		updatedItem.Tags,
		updatedItem.Image,
		updatedItem.Seasonal,
		id,
	).Scan(&returnedID)
	
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			http.Error(w, "Item with this name already exists", http.StatusConflict)
			return
		}
		log.Printf("Failed to update item: %v", err)
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	
	w.WriteHeader(http.StatusOK)
}

func (h *SpeisekarteHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	cmdTag, err := h.DB.Exec(context.Background(),
		`DELETE FROM speisekarte WHERE id = $1`, id)
	if err != nil {
		log.Printf("Failed to delete item: %v", err)
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	
	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
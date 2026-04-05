package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gomisroca/gasthaus-backend/internal"
	"github.com/gomisroca/gasthaus-backend/models"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	cacheMu             sync.RWMutex
	menuCache           = make(map[string][]models.SpeisekarteItem)
	cacheTimes          = make(map[string]time.Time)
	categoriesCache     []string
	categoriesCacheTime time.Time
	cacheDuration       = 5 * time.Minute
)

type SpeisekarteHandler struct {
	DB *pgxpool.Pool
}

func invalidateCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	menuCache = make(map[string][]models.SpeisekarteItem)
	cacheTimes = make(map[string]time.Time)
	categoriesCache = nil
	categoriesCacheTime = time.Time{}
}

func (h *SpeisekarteHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	cacheMu.RLock()
	cached := categoriesCache
	cacheTime := categoriesCacheTime
	cacheMu.RUnlock()

	if time.Since(cacheTime) < cacheDuration && cached != nil {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cached); err != nil {
			log.Printf("Error encoding categories response: %v", err)
		}
		return
	}

	rows, err := h.DB.Query(r.Context(), "SELECT DISTINCT unnest(categories) AS category FROM speisekarte")
	if err != nil {
		log.Printf("Database query failed: %v", err)
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Failed to read categories", http.StatusInternalServerError)
		return
	}

	cacheMu.Lock()
	categoriesCache = categories
	categoriesCacheTime = time.Now()
	cacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		log.Printf("Error encoding categories response: %v", err)
	}
}

func (h *SpeisekarteHandler) GetUniqueItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, name, description, price_cents, categories, ingredients, tags, image, seasonal, created_at, updated_at
		FROM speisekarte
		WHERE id = $1
	`

	var item models.SpeisekarteItem
	err := h.DB.QueryRow(r.Context(), query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.PriceCents,
		&item.Categories,
		&item.Ingredients,
		&item.Tags,
		&item.Image,
		&item.Seasonal,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to fetch item: %v", err)
		http.Error(w, "Failed to fetch item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("Error encoding item response: %v", err)
	}
}

func (h *SpeisekarteHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	cacheMu.RLock()
	data, ok := menuCache[category]
	cacheTime := cacheTimes[category]
	cacheMu.RUnlock()

	if ok && time.Since(cacheTime) < cacheDuration {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding items response: %v", err)
		}
		return
	}

	var rows pgx.Rows
	var err error

	if category == "" {
		rows, err = h.DB.Query(r.Context(),
			`SELECT id, name, description, price_cents, categories, ingredients, tags, image, seasonal, created_at, updated_at
			 FROM speisekarte`)
	} else {
		rows, err = h.DB.Query(r.Context(),
			`SELECT id, name, description, price_cents, categories, ingredients, tags, image, seasonal, created_at, updated_at
			 FROM speisekarte WHERE $1 = ANY(categories)`, category)
	}

	if err != nil {
		log.Printf("Database query failed: %v", err)
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.SpeisekarteItem
	for rows.Next() {
		var item models.SpeisekarteItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.PriceCents,
			&item.Categories,
			&item.Ingredients,
			&item.Tags,
			&item.Image,
			&item.Seasonal,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		http.Error(w, "Failed to read items", http.StatusInternalServerError)
		return
	}

	cacheMu.Lock()
	menuCache[category] = items
	cacheTimes[category] = time.Now()
	cacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Error encoding items response: %v", err)
	}
}

func (h *SpeisekarteHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price_cents")
	categories := r.Form["categories"]
	ingredients := r.Form["ingredients"]
	tags := r.Form["tags"]
	seasonal := r.FormValue("seasonal") == "true"

	if name == "" || priceStr == "" || len(categories) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	priceCents, err := strconv.Atoi(priceStr)
	if err != nil || priceCents < 0 {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageURL, err := internal.UploadToSupabase(r.Context(), file, handler)
	if err != nil {
		log.Printf("Image upload failed: %v", err)
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return
	}

	query := `INSERT INTO speisekarte (name, description, price_cents, categories, ingredients, tags, image, seasonal)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = h.DB.Exec(
		r.Context(),
		query,
		name,
		description,
		priceCents,
		categories,
		ingredients,
		tags,
		imageURL,
		seasonal,
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

	invalidateCache()
	w.WriteHeader(http.StatusCreated)
}

func (h *SpeisekarteHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	priceStr := r.FormValue("price_cents")
	categories := r.Form["categories"]
	ingredients := r.Form["ingredients"]
	tags := r.Form["tags"]
	seasonal := r.FormValue("seasonal") == "true"

	if name == "" || priceStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	priceCents, err := strconv.Atoi(priceStr)
	if err != nil || priceCents < 0 {
		http.Error(w, "Invalid price", http.StatusBadRequest)
		return
	}

	var imageURL string
	err = h.DB.QueryRow(r.Context(), `SELECT image FROM speisekarte WHERE id = $1`, id).Scan(&imageURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to fetch existing item: %v", err)
		http.Error(w, "Failed to fetch item", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("image")
	if err == nil && file != nil {
		defer file.Close()
		uploadedURL, uploadErr := internal.UploadToSupabase(r.Context(), file, handler)
		if uploadErr != nil {
			log.Printf("Image upload failed: %v", uploadErr)
			http.Error(w, "Image upload failed", http.StatusInternalServerError)
			return
		}
		imageURL = uploadedURL
	}

	query := `
		UPDATE speisekarte
		SET name = $1,
			description = $2,
			price_cents = $3,
			categories = $4,
			ingredients = $5,
			tags = $6,
			image = $7,
			seasonal = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING id
	`

	var returnedID string
	err = h.DB.QueryRow(
		r.Context(),
		query,
		name,
		description,
		priceCents,
		categories,
		ingredients,
		tags,
		imageURL,
		seasonal,
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

	invalidateCache()
	w.WriteHeader(http.StatusOK)
}

func (h *SpeisekarteHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	cmdTag, err := h.DB.Exec(r.Context(), `DELETE FROM speisekarte WHERE id = $1`, id)
	if err != nil {
		log.Printf("Failed to delete item: %v", err)
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	invalidateCache()
	w.WriteHeader(http.StatusOK)
}
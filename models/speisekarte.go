package models

import "time"

type SpeisekarteItem struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Categories  []string  `json:"categories" db:"categories"`
	Ingredients []string  `json:"ingredients" db:"ingredients"`
	Tags        []string  `json:"tags" db:"tags"`
	PriceCents  int       `json:"price_cents" db:"price_cents"`
	Image       *string   `json:"image" db:"image"`
	Seasonal    bool      `json:"seasonal" db:"seasonal"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
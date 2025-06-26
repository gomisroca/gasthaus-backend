package models

type SpeisekarteItem struct {
	ID   string    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`
	Categories []string `json:"categories" db:"categories"`
	Tags []string `json:"tags" db:"tags"`
	Price float32 `json:"price" db:"price"`
	Image *string `json:"image" db:"image"`
	Seasonal bool `json:"seasonal" db:"seasonal"`
}

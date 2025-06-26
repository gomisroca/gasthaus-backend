package models

type User struct {
	ID   string    `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	PasswordHash string `json:"password_hash" db:"password_hash"`
}

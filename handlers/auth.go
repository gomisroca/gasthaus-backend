package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gomisroca/gasthaus-backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)


type AuthHandler struct {
	DB *pgxpool.Pool
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Look up user by email
	var user models.User
	err := h.DB.QueryRow(
		context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email=$1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("JWT_SECRET not set in environment")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(loginResponse{Token: tokenString})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// This handler would typically handle token refresh logic
	// For simplicity, we will just demand a token and return a success message
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Here you would typically validate the token and refresh it
	if token != "valid-token" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// If the token is valid, we can proceed with the refresh
	// In a real application, you would generate a new token and return it
	// For this example, we will just return a new random token
	newToken := "new-valid-token"
	fmt.Fprintf(w, "Token refreshed successfully! New token: %s\n", newToken)
}
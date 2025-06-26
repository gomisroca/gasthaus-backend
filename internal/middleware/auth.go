package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Token is not present in the request, return unauthorized
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Check if the token is in the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "JWT_SECRET not set in environment", http.StatusInternalServerError)
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Check signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		// If something went wrong with the parsing, return unauthorized
		if err != nil  || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, continue to the next handler
		next.ServeHTTP(w, r)
	})
}
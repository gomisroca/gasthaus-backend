package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomisroca/gasthaus-backend/internal"
	"github.com/gomisroca/gasthaus-backend/routes"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func healthCheckHandler(dbpool *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := dbpool.Ping(r.Context()); err != nil {
            http.Error(w, "DB unavailable", http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "OK")
    }
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	if err := internal.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	dbpool, err := internal.SetupDB()
	if err != nil {
		log.Fatalf("Failed to set up DB: %v", err)
	}
	defer dbpool.Close()
	fmt.Println("DB connected successfully")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))
	
	r.HandleFunc("/", healthCheckHandler(dbpool)).Methods("GET")
	routes.RegisterAuthRoutes(r, dbpool, jwtSecret)
	routes.RegisterSpeisekarteRoutes(r, dbpool, jwtSecret)

	// CORS setup
	origin := os.Getenv("FRONTEND_ORIGIN")
	if origin == "" {
		log.Fatal("FRONTEND_ORIGIN environment variable is not set")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(r)

	// Create http.Server with your router and config
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr: ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt or terminate signal from OS
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine so it doesn’t block
	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Block until we receive signal
	<-stopChan
	log.Println("Shutdown signal received, shutting down server gracefully...")

	// Create a deadline to wait for current operations to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server gracefully — stops accepting new requests,
	// waits for ongoing requests, or timeout after 5 seconds
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server exited properly")
}

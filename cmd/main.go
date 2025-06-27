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
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Health check successful!\n")
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	internal.RunMigrations()

	dbpool, err := internal.SetupDB()
	if err != nil {
		log.Fatalf("Failed to set up DB: %v", err)
	}
	defer dbpool.Close()
	fmt.Println("DB connected successfully")

	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("static/"))
	r.Handle("/static/", http.StripPrefix("/static/", fs))

	r.HandleFunc("/", healthCheckHandler).Methods("GET")
	routes.RegisterAuthRoutes(r, dbpool)
	routes.RegisterSpeisekarteRoutes(r, dbpool)

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{os.Getenv("FRONTEND_ORIGIN")},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
	})

	handler := c.Handler(r)

	// Create http.Server with your router and config
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
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

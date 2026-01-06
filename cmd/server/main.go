package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/stripsior/loc.api/internal/api"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cacheTTL := 1 * time.Hour
	if ttlEnv := os.Getenv("CACHE_TTL_MINUTES"); ttlEnv != "" {
		if minutes, err := strconv.Atoi(ttlEnv); err == nil {
			cacheTTL = time.Duration(minutes) * time.Minute
		}
	}

	// Configure timeouts for large repositories
	readTimeout := 5 * time.Minute
	if rtEnv := os.Getenv("HTTP_READ_TIMEOUT_SECONDS"); rtEnv != "" {
		if seconds, err := strconv.Atoi(rtEnv); err == nil {
			readTimeout = time.Duration(seconds) * time.Second
		}
	}

	writeTimeout := 10 * time.Minute
	if wtEnv := os.Getenv("HTTP_WRITE_TIMEOUT_SECONDS"); wtEnv != "" {
		if seconds, err := strconv.Atoi(wtEnv); err == nil {
			writeTimeout = time.Duration(seconds) * time.Second
		}
	}

	idleTimeout := 2 * time.Minute
	if itEnv := os.Getenv("HTTP_IDLE_TIMEOUT_SECONDS"); itEnv != "" {
		if seconds, err := strconv.Atoi(itEnv); err == nil {
			idleTimeout = time.Duration(seconds) * time.Second
		}
	}

	router := api.SetupRouter(cacheTTL)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting LOC Counter API on %s", addr)
	log.Printf("Cache TTL: %s", cacheTTL)
	log.Printf("HTTP Timeouts - Read: %s, Write: %s, Idle: %s", readTimeout, writeTimeout, idleTimeout)
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("Cache stats: http://localhost%s/cache/stats", addr)
	log.Printf("Analyze endpoint: POST http://localhost%s/api/analyze", addr)

	// Use custom HTTP server with configured timeouts
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

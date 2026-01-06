package main

import (
	"fmt"
	"log"
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

	router := api.SetupRouter(cacheTTL)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting LOC Counter API on %s", addr)
	log.Printf("Cache TTL: %s", cacheTTL)
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("Cache stats: http://localhost%s/cache/stats", addr)
	log.Printf("Analyze endpoint: POST http://localhost%s/api/analyze", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

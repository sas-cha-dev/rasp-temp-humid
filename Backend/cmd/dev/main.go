package main

import (
	"BeRoHuTe/internal/handler"
	"BeRoHuTe/internal/repository"
	"BeRoHuTe/internal/sensor"
	"context"
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Load configuration from environment
	readInterval := getEnvInt("READ_INTERVAL", 60) // default 60 seconds
	dbPath := getEnv("DB_PATH", "./data.db")
	port := getEnv("PORT", "8080")
	templateDir := getEnv("TEMPLATE_DIR", "./web")

	// init db
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	///////////////////////// Repos /////////////////////////

	// Initialize repository
	repo, err := repository.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Init Button repo
	btnRepo, err := repository.NewButtonRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize button repository: %v", err)
	}

	// Initialize sensors
	sensorService := sensor.NewDummyService()
	btnService := sensor.NewDummyButtonService(24)

	///////////////////////// Applications /////////////////////////

	ctx := context.Background()
	defer ctx.Done()

	// read dht sensors
	dhtApp := sensor.NewSensorService(sensorService, repo)
	dhtApp.Start(ctx)
	defer dhtApp.Stop()

	btnApp, err := sensor.NewButtonApp(btnService, btnRepo)
	if err != nil {
		log.Fatalf("Failed to initialize button application: %v", err)
	}
	if err := btnApp.Start(ctx); err != nil {
		log.Fatalf("Failed to start button application: %v", err)
	}

	// Initialize HTTP handler
	h, err := handler.New(repo, templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize handler: %v", err)
	}

	// Setup routes
	http.HandleFunc("/", h.ServeIndex)
	http.HandleFunc("/api/data", h.ServeAPI)

	// Start server
	log.Printf("Starting server on port %s, reading sensors every %d seconds", port, readInterval)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

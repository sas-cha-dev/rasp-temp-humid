package main

import (
	"BeRoHuTe/internal/handler"
	"BeRoHuTe/internal/repository"
	"BeRoHuTe/internal/sensor"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
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
	buttonService := sensor.NewDummyButtonService(10)

	// Start background sensor reading
	go readSensors(sensorService, repo, readInterval)

	var startsAt, endsAt time.Time
	_ = buttonService.OnPush(func(state sensor.ButtonState) error {
		startsAt = time.Now()
		log.Printf("button pushed at %v", startsAt)
		return nil
	})
	_ = buttonService.OnRelease(func(state sensor.ButtonState) error {
		endsAt = time.Now()
		if startsAt.IsZero() {
			log.Printf("button released but never pushed at %v", endsAt)
			return nil
		}
		err := btnRepo.Save(10, startsAt, endsAt)
		if err != nil {
			return err
		}
		startsAt = time.Time{}
		endsAt = time.Time{}
		log.Println("Saved button ", startsAt, endsAt)
		return nil
	})
	err = buttonService.Start(time.Minute / 2)
	if err != nil {
		log.Fatalf("Failed to start button service: %v", err)
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

func readSensors(sensorService sensor.Service, repo *repository.Repository, intervalSeconds int) {
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	// Read immediately on start
	performReading(sensorService, repo)

	// Then read at intervals
	for range ticker.C {
		performReading(sensorService, repo)
	}
}

func performReading(sensorService sensor.Service, repo *repository.Repository) {
	readings, err := sensorService.ReadAllSensors()
	if err != nil {
		log.Printf("Error reading sensors: %v", err)
		return
	}

	for _, reading := range readings {
		err := repo.Save(reading.SensorID, reading.Temperature, reading.Humidity, reading.Timestamp)
		if err != nil {
			log.Printf("Error saving reading for sensor %d: %v", reading.SensorID, err)
		} else {
			log.Printf("Saved: Sensor %d - Temp: %.1f°C, Humidity: %.1f%%, Time: %s",
				reading.SensorID, reading.Temperature, reading.Humidity, reading.Timestamp.Format("15:04:05"))
		}
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

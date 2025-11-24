package main

import (
	"BeRoHuTe/internal/handler"
	"BeRoHuTe/internal/repository"
	"BeRoHuTe/internal/rpi"
	"BeRoHuTe/internal/sensor"
	"BeRoHuTe/internal/weather"
	"context"
	"database/sql"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

	weatherReadInterval := getEnvInt("WEATHER_READ_INTERVAL_MIN", 30) // in minutes
	openWeatherApiKey := getEnv("OPEN_WEATHER_API_KEY", "")
	locationCoords := getEnv("LOCATION_COORDS", "")

	var locationLon, locationLat float64
	if strings.TrimSpace(locationCoords) != "" {
		lonLat := strings.Split(locationCoords, ",")

		var err error
		locationLat, err = strconv.ParseFloat(lonLat[0], 64)
		if err != nil {
			log.Fatal(err)
		}
		locationLon, err = strconv.ParseFloat(lonLat[1], 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	// init db
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})

	///////////////////////// Repos /////////////////////////

	// Initialize repositories
	repo, err := repository.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	btnRepo, err := repository.NewButtonRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize button repository: %v", err)
	}
	weatherRepo, err := repository.NewWeatherRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize weather repository: %v", err)
	}

	// Initialize sensors
	sensorService := sensor.NewDHTSensors(rdb)
	btnService, err := rpi.NewRealButtonService(24)
	if err != nil {
		log.Fatalf("Failed to initialize button service: %v", err)
	}
	weatherService := weather.NewOpenWeatherService(
		openWeatherApiKey,
		locationLat,
		locationLon)

	///////////////////////// Applications /////////////////////////

	ctx := context.Background()
	defer ctx.Done()

	// read dht sensors
	dhtApp := sensor.NewSensorService(time.Duration(readInterval)*time.Second, sensorService, repo)
	dhtApp.Start(ctx)
	defer dhtApp.Stop()

	btnApp, err := sensor.NewButtonApp(btnService, btnRepo)
	if err != nil {
		log.Fatalf("Failed to initialize button application: %v", err)
	}
	if err := btnApp.Start(ctx); err != nil {
		log.Fatalf("Failed to start button application: %v", err)
	}

	weatherApp := weather.NewApp(weatherService, weatherRepo)
	weatherApp.Start(ctx, time.Duration(weatherReadInterval)*time.Minute)
	defer weatherApp.Stop()

	// Initialize HTTP handler
	h, err := handler.New(repo, templateDir, btnRepo, weatherRepo)
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

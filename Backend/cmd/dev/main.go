package main

import (
	"BeRoHuTe/config"
	"BeRoHuTe/internal/buttons"
	"BeRoHuTe/internal/data_clean"
	"BeRoHuTe/internal/handler"
	"BeRoHuTe/internal/sensor"
	"BeRoHuTe/internal/weather"
	"BeRoHuTe/util"
	"context"
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
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
	readInterval := util.GetEnvInt("READ_INTERVAL", 60) // default 60 seconds
	dbPath := util.GetEnv("DB_PATH", "./data.db")
	port := util.GetEnv("PORT", "8080")
	templateDir := util.GetEnv("TEMPLATE_DIR", "./web")

	weatherReadInterval := util.GetEnvInt("WEATHER_READ_INTERVAL_MIN", 30) // in minutes
	openWeatherApiKey := util.GetEnv("OPEN_WEATHER_API_KEY", "")
	locationCoords := util.GetEnv("LOCATION_COORDS", "")

	progArgs, err := config.GetProgramArgs()
	if err != nil {
		log.Fatal(err)
	}

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

	///////////////////////// Repos /////////////////////////

	// Initialize repositories
	repo, err := sensor.New(db)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	btnRepo, err := buttons.NewButtonRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize button repository: %v", err)
	}
	weatherRepo, err := weather.NewWeatherRepository(db)
	if err != nil {
		log.Fatalf("Failed to initialize weather repository: %v", err)
	}

	// Initialize sensors
	sensorService := sensor.NewDummyService()
	btnService := buttons.NewDummyService(24)
	weatherService := weather.NewOpenWeatherService(
		openWeatherApiKey,
		locationLat,
		locationLon)

	///////////////////////// Applications /////////////////////////

	ctx := context.Background()
	defer ctx.Done()

	dhtApp := sensor.NewApp(time.Duration(readInterval)*time.Second, sensorService, repo)
	dhtApp.Start(ctx, true)
	defer dhtApp.Stop()

	btnApp, err := buttons.NewButtonApp(btnService, btnRepo)
	if err != nil {
		log.Fatalf("Failed to initialize button application: %v", err)
	}
	if err := btnApp.Start(ctx); err != nil {
		log.Fatalf("Failed to start button application: %v", err)
	}

	weatherApp := weather.NewApp(weatherService, weatherRepo)
	weatherApp.Start(ctx, time.Duration(weatherReadInterval)*time.Minute)
	defer weatherApp.Stop()

	if progArgs.Cleanup {
		dataCleanUp, err := data_clean.NewApp(btnRepo, repo,
			data_clean.WithBeforeCleanUp(func() error {
				dhtApp.Stop()
				if err := btnApp.Stop(); err != nil {
					return err
				}
				return nil
			}),
			data_clean.WithAfterCleanUp(func() error {
				dhtApp.Start(ctx, false)
				if err := btnApp.Start(ctx); err != nil {
					return err
				}
				return nil
			}),
		)
		if err != nil {
			log.Fatalf("Failed to initialize data clean: %v", err)
		}
		dataCleanUp.Start(ctx, 24*time.Hour, true)
		defer dataCleanUp.Stop()
	}

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

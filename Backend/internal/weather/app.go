package weather

import (
	"BeRoHuTe/internal/contracts"
	"context"
	"log"
	"time"
)

type App struct {
	service Service
	repo    WeatherRepository
	start   bool
	stop    chan bool
}

func NewApp(s Service, r WeatherRepository) *App {
	return &App{
		service: s,
		repo:    r,
	}
}

func (a *App) Start(ctx context.Context, dur time.Duration) {
	if a.start == true {
		return
	}
	a.start = true
	a.stop = make(chan bool)

	ticker := time.NewTicker(1 * time.Millisecond) // directly call it

	go func() {
		defer ticker.Stop()

		errors := 0

		for range ticker.C {
			select {
			case <-a.stop:
			case <-ctx.Done():
				return
			default:
			}

			if errors > 5 { // wait a longer period
				ticker.Reset(dur * 2)
				log.Println("Weather app 'stopped' due to timeout")
				continue
			}

			err := a.fetchAndStoreCurrentWeatherDetails()
			if err != nil {
				log.Println("WeatherApp error: ", err)
				errors++
				ticker.Reset(2 * time.Second)
				continue
			}

			errors = 0
			ticker.Reset(dur)
		}
	}()
}

func (a *App) fetchAndStoreCurrentWeatherDetails() error {
	details, err := a.service.GetCurrentWeatherDetails()
	if err != nil {
		return err
	}

	// save to repository
	err = a.repo.Save(contracts.WeatherData{
		Time:        time.Unix(details.Timestamp, 0),
		Name:        "Home",
		Latitude:    details.Latitude,
		Longitude:   details.Longitude,
		Temperature: details.Temperature,
		Humidity:    details.Humidity,
		FeelsLike:   details.FeelsLike,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.stop <- true
}

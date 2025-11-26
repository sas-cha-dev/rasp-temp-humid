package sensor

import (
	"context"
	"log"
	"time"
)

type DHTApp struct {
	service        Service
	repo           *Repository
	stop           chan bool
	lastTimestamps map[int]time.Time
	interval       time.Duration
}

func NewApp(readInterval time.Duration, sensorService Service, repo *Repository) *DHTApp {
	return &DHTApp{
		service:        sensorService,
		repo:           repo,
		stop:           make(chan bool),
		lastTimestamps: map[int]time.Time{},
		interval:       readInterval,
	}
}

func (sensorApp *DHTApp) Start(ctx context.Context, execDirectly bool) {
	sensorApp.stop = make(chan bool)
	startInterval := sensorApp.interval
	if execDirectly {
		startInterval = time.Millisecond
	}
	ticker := time.NewTicker(startInterval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sensorApp.stop:
				return
			case <-ticker.C:
				sensorApp.performReading()
				ticker.Reset(sensorApp.interval)
			default:
			}
		}
	}()
}

func (sensorApp *DHTApp) performReading() {
	readings, err := sensorApp.service.ReadAllSensors()
	if err != nil {
		log.Printf("Error reading sensors: %v", err)
		return
	}

	for _, reading := range readings {
		lastTimestamp, ok := sensorApp.lastTimestamps[reading.SensorID]
		if ok && (lastTimestamp.After(reading.Timestamp) ||
			lastTimestamp.Equal(reading.Timestamp)) {
			continue
		}

		sensorApp.lastTimestamps[reading.SensorID] = reading.Timestamp

		// save to repository
		err := sensorApp.repo.Save(reading.SensorID, reading.Temperature, reading.Humidity, reading.Timestamp)
		if err != nil {
			log.Printf("Error saving reading for sensor %d: %v", reading.SensorID, err)
		} else {
			log.Printf("Saved: Sensor %d - Temp: %.1f°C, Humidity: %.1f%%, Time: %s",
				reading.SensorID, reading.Temperature, reading.Humidity, reading.Timestamp.Format("15:04:05"))
		}
	}
}

func (sensorApp *DHTApp) Stop() {
	sensorApp.stop <- true
}

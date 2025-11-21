package sensor

import (
	"BeRoHuTe/internal/repository"
	"context"
	"log"
	"time"
)

type SensorApp struct {
	service        Service
	repo           *repository.SensorRepository
	stop           chan bool
	lastTimestamps map[int]time.Time
}

func NewSensorService(sensorService Service, repo *repository.SensorRepository) *SensorApp {
	return &SensorApp{
		service: sensorService,
		repo:    repo,
		stop:    make(chan bool),
	}
}

func (sensorApp *SensorApp) Start(ctx context.Context) {
	sensorApp.stop = make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sensorApp.stop:
			return
		default:
		}

		sensorApp.readSensors()
	}()
}

func (sensorApp *SensorApp) readSensors() {
	ticker := time.NewTimer(time.Duration(60) * time.Second)
	defer ticker.Stop()

	sensorApp.performReading()

	for range ticker.C {
		sensorApp.performReading()
	}
}

func (sensorApp *SensorApp) performReading() {
	readings, err := sensorApp.service.ReadAllSensors()
	if err != nil {
		log.Printf("Error reading sensors: %v", err)
		return
	}

	for _, reading := range readings {
		lastTimestamp, ok := sensorApp.lastTimestamps[reading.SensorID]
		if ok && (lastTimestamp.Before(reading.Timestamp) ||
			lastTimestamp.Equal(reading.Timestamp)) {
			continue
		}

		sensorApp.lastTimestamps[reading.SensorID] = reading.Timestamp

		// save to repository
		err := sensorApp.repo.Save(reading.SensorID, reading.TemperatureC, reading.Humidity, reading.Timestamp)
		if err != nil {
			log.Printf("Error saving reading for sensor %d: %v", reading.SensorID, err)
		} else {
			log.Printf("Saved: Sensor %d - Temp: %.1f°C, Humidity: %.1f%%, Time: %s",
				reading.SensorID, reading.TemperatureC, reading.Humidity, reading.Timestamp.Format("15:04:05"))
		}
	}
}

func (sensorApp *SensorApp) Stop() {

}

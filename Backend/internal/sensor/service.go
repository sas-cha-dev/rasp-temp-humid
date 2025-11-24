package sensor

import (
	"fmt"
	"math"
	"time"
)

// Reading represents a single sensor reading
type Reading struct {
	SensorID    int
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Timestamp   time.Time `json:"timestamp"`
}

// Service defines the interface for reading sensor data
type Service interface {
	// ReadSensor reads data from a specific sensor (1 or 2)
	ReadSensor(sensorID int) (*Reading, error)
	// ReadAllSensors reads data from all sensors
	ReadAllSensors() ([]*Reading, error)
}

// DummyService simulates sensor readings with values that change over time
type DummyService struct {
	startTime time.Time
}

// NewDummyService creates a new dummy sensor service
func NewDummyService() *DummyService {
	return &DummyService{
		startTime: time.Now(),
	}
}

// ReadSensor reads simulated data from a specific sensor
func (d *DummyService) ReadSensor(sensorID int) (*Reading, error) {
	if sensorID < 1 || sensorID > 2 {
		return nil, fmt.Errorf("invalid sensor ID: %d", sensorID)
	}

	now := time.Now()
	elapsed := now.Sub(d.startTime).Seconds()

	// Generate temperature and humidity based on sine waves for variation
	// Sensor 1: 20-24°C, 45-55% humidity
	// Sensor 2: 19-23°C, 50-60% humidity
	var temp, humidity float64

	if sensorID == 1 {
		temp = 22.0 + 2.0*math.Sin(elapsed/10.0)
		humidity = 50.0 + 5.0*math.Cos(elapsed/15.0)
	} else {
		temp = 21.0 + 2.0*math.Sin(elapsed/12.0)
		humidity = 55.0 + 5.0*math.Cos(elapsed/18.0)
	}

	return &Reading{
		SensorID:    sensorID,
		Temperature: math.Round(temp*10) / 10,     // Round to 1 decimal
		Humidity:    math.Round(humidity*10) / 10, // Round to 1 decimal
		Timestamp:   now,
	}, nil
}

// ReadAllSensors reads simulated data from all sensors
func (d *DummyService) ReadAllSensors() ([]*Reading, error) {
	readings := make([]*Reading, 0, 2)

	for i := 1; i <= 2; i++ {
		reading, err := d.ReadSensor(i)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

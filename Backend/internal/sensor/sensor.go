package sensor

import "time"

// Reading represents a single sensor reading
type Reading struct {
	SensorID    int       `json:"sensor_id"`
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

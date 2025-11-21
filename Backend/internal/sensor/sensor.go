package sensor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// Reading represents a single sensor reading
type Reading struct {
	SensorID     int
	TemperatureC float64   `json:"temperature_c"`
	TemperatureF float64   `json:"temperature_f"`
	Humidity     float64   `json:"humidity"`
	Timestamp    time.Time `json:"timestamp"`
}

// Service defines the interface for reading sensor data
type Service interface {
	// ReadSensor reads data from a specific sensor (1 or 2)
	ReadSensor(sensorID int) (*Reading, error)
	// ReadAllSensors reads data from all sensors
	ReadAllSensors() ([]*Reading, error)
}

var allSensorIDs = []int{1, 2}

type DHTSensors struct {
	rdb *redis.Client
}

func NewDHTSensors(re *redis.Client) *DHTSensors {
	return &DHTSensors{rdb: re}
}

func (D DHTSensors) ReadSensor(sensorID int) (*Reading, error) {
	// read from redis 'sensor<id>'
	res, err := D.rdb.Get(context.Background(), fmt.
		Sprintf("sensor:%d", sensorID)).
		Result()
	if err != nil {
		return nil, err
	}

	var reading Reading

	if err := json.Unmarshal([]byte(res), &reading); err != nil {
		return nil, err
	}

	reading.SensorID = sensorID

	return &reading, nil
}

func (D DHTSensors) ReadAllSensors() ([]*Reading, error) {
	res := make([]*Reading, 0, len(allSensorIDs))

	for _, sensorID := range allSensorIDs {
		reading, err := D.ReadSensor(sensorID)
		if err != nil {
			return nil, err
		}
		res = append(res, reading)
	}

	return res, nil
}

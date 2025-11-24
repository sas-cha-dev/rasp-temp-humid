package sensor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type sensorReading struct {
	TemperatureC float64 `json:"temperature_c"`
	TemperatureF float64 `json:"temperature_f"`
	Humidity     float64 `json:"humidity"`
	Timestamp    float64 `json:"timestamp"`
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
		Sprintf("sensor%d", sensorID)).
		Result()
	if err != nil {
		return nil, err
	}

	var reading sensorReading

	if err := json.Unmarshal([]byte(res), &reading); err != nil {
		return nil, err
	}

	r := Reading{
		SensorID:    sensorID,
		Temperature: reading.TemperatureC,
		Humidity:    reading.Humidity,
		Timestamp:   time.Unix(int64(reading.Timestamp), 0),
	}

	return &r, nil
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

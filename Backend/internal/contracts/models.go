package contracts

import "time"

type SensorReading struct {
	ID          int64     `json:"id"`
	SensorID    int       `json:"sensor_id"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Timestamp   time.Time `json:"timestamp"`
}

type ButtonReading struct {
	ID       int64     `json:"id"`
	ButtonID int       `json:"button_id"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
}

type WeatherData struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Time        time.Time `json:"created_at"`
	Latitude    float32   `json:"latitude"`
	Longitude   float32   `json:"longitude"`
	Temperature float32   `json:"temperature"`
	Humidity    float32   `json:"humidity"`
	FeelsLike   float32   `json:"feels_like"`
}

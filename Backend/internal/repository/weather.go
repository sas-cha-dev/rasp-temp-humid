package repository

import (
	"database/sql"
	"log"
	"time"
)

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

type WeatherRepository interface {
	Save(weather WeatherData) error
	GetLatest() ([]*WeatherData, error)
}

type weatherRepository struct {
	db *sql.DB
}

func NewWeatherRepository(db *sql.DB) (WeatherRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &weatherRepository{db: db}
	if err := repo.createTable(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (w *weatherRepository) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS weather_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    latitude REAL,
    longitude REAL,
    temperature REAL,
    humidity REAL,
    feels_like REAL)`
	_, err := w.db.Exec(query)
	return err
}

func (w *weatherRepository) Save(weather WeatherData) error {
	query := `INSERT INTO weather_data (time, name, latitude, longitude, temperature, humidity, feels_like) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := w.db.Exec(query, weather.Time, weather.Name, weather.Latitude, weather.Longitude, weather.Temperature, weather.Humidity, weather.FeelsLike)
	return err
}

func (w *weatherRepository) GetLatest() ([]*WeatherData, error) {
	query := `SELECT id, time, name, latitude, longitude, temperature,humidity,feels_like FROM weather_data ORDER BY time DESC LIMIT 1`
	rows, err := w.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return w.queryReadings(query)
}

func (w *weatherRepository) queryReadings(query string, args ...interface{}) ([]*WeatherData, error) {
	rows, err := w.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*WeatherData
	for rows.Next() {
		var datum WeatherData
		err := rows.Scan(&datum.ID, &datum.Time, &datum.Name, &datum.Latitude, &datum.Longitude, &datum.Temperature, &datum.Humidity, &datum.FeelsLike)
		if err != nil {
			return nil, err
		}
		data = append(data, &datum)
	}

	return data, rows.Err()
}

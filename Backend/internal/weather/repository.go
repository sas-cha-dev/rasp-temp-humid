package weather

import (
	"BeRoHuTe/internal/contracts"
	"database/sql"
	"log"
)

type WeatherRepository interface {
	Save(weather contracts.WeatherData) error
	GetLatest() ([]*contracts.WeatherData, error)
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

func (w *weatherRepository) Save(weather contracts.WeatherData) error {
	query := `INSERT INTO weather_data (time, name, latitude, longitude, temperature, humidity, feels_like) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := w.db.Exec(query, weather.Time, weather.Name, weather.Latitude, weather.Longitude, weather.Temperature, weather.Humidity, weather.FeelsLike)
	return err
}

func (w *weatherRepository) GetLatest() ([]*contracts.WeatherData, error) {
	query := `SELECT id, time, name, latitude, longitude, temperature,humidity,feels_like FROM weather_data ORDER BY time DESC LIMIT 1`
	rows, err := w.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return w.queryReadings(query)
}

func (w *weatherRepository) queryReadings(query string, args ...interface{}) ([]*contracts.WeatherData, error) {
	rows, err := w.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*contracts.WeatherData
	for rows.Next() {
		var datum contracts.WeatherData
		err := rows.Scan(&datum.ID, &datum.Time, &datum.Name, &datum.Latitude, &datum.Longitude, &datum.Temperature, &datum.Humidity, &datum.FeelsLike)
		if err != nil {
			return nil, err
		}
		data = append(data, &datum)
	}

	return data, rows.Err()
}

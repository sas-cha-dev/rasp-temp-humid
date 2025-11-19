package repository

import (
	"database/sql"
	"time"
)

type Reading struct {
	ID          int64     `json:"id"`
	SensorID    int       `json:"sensor_id"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Timestamp   time.Time `json:"timestamp"`
}

type Repository struct {
	db *sql.DB
}

// New creates a new repository and initializes the database
func New(db *sql.DB) (*Repository, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &Repository{db: db}
	if err := repo.createTable(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS readings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sensor_id INTEGER NOT NULL,
		temperature REAL NOT NULL,
		humidity REAL NOT NULL,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sensor_timestamp ON readings(sensor_id, timestamp);
	`
	_, err := r.db.Exec(query)
	return err
}

// Save stores a new reading
func (r *Repository) Save(sensorID int, temperature, humidity float64, timestamp time.Time) error {
	query := `INSERT INTO readings (sensor_id, temperature, humidity, timestamp) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, sensorID, temperature, humidity, timestamp)
	return err
}

// GetLatest returns the latest reading for each sensor
func (r *Repository) GetLatest() ([]*Reading, error) {
	query := `
	SELECT id, sensor_id, temperature, humidity, timestamp
	FROM readings
	WHERE (sensor_id, timestamp) IN (
		SELECT sensor_id, MAX(timestamp)
		FROM readings
		GROUP BY sensor_id
	)
	ORDER BY sensor_id
	`
	return r.queryReadings(query)
}

// GetLastN returns the last N readings for all sensors
func (r *Repository) GetLastN(n int) ([]*Reading, error) {
	query := `
	SELECT id, sensor_id, temperature, humidity, timestamp
	FROM readings
	ORDER BY timestamp DESC
	LIMIT ?
	`
	return r.queryReadings(query, n)
}

// GetAverageLastHour returns average temperature and humidity for each sensor in the last hour
func (r *Repository) GetAverageLastHour() (map[int]map[string]float64, error) {
	query := `
	SELECT sensor_id, AVG(temperature) as avg_temp, AVG(humidity) as avg_humidity
	FROM readings
	WHERE timestamp >= datetime('now', '-1 hour')
	GROUP BY sensor_id
	`
	return r.queryAverages(query)
}

// GetAverageToday returns average temperature and humidity for each sensor today
func (r *Repository) GetAverageToday() (map[int]map[string]float64, error) {
	query := `
	SELECT sensor_id, AVG(temperature) as avg_temp, AVG(humidity) as avg_humidity
	FROM readings
	WHERE timestamp > date('now') AND timestamp <= date('now', '+1 day')
	GROUP BY sensor_id
	`
	return r.queryAverages(query)
}

// GetAverageThisWeek returns average temperature and humidity for each sensor this week
func (r *Repository) GetAverageThisWeek() (map[int]map[string]float64, error) {
	query := `
	SELECT sensor_id, AVG(temperature) as avg_temp, AVG(humidity) as avg_humidity
	FROM readings
	WHERE timestamp >= datetime('now', '-7 days')
	GROUP BY sensor_id
	`
	return r.queryAverages(query)
}

func (r *Repository) queryReadings(query string, args ...interface{}) ([]*Reading, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*Reading
	for rows.Next() {
		var reading Reading
		err := rows.Scan(&reading.ID, &reading.SensorID, &reading.Temperature, &reading.Humidity, &reading.Timestamp)
		if err != nil {
			return nil, err
		}
		readings = append(readings, &reading)
	}

	return readings, rows.Err()
}

func (r *Repository) queryAverages(query string) (map[int]map[string]float64, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]map[string]float64)
	for rows.Next() {
		var sensorID int
		var avgTemp, avgHumidity float64
		err := rows.Scan(&sensorID, &avgTemp, &avgHumidity)
		if err != nil {
			return nil, err
		}
		result[sensorID] = map[string]float64{
			"temperature": avgTemp,
			"humidity":    avgHumidity,
		}
	}

	return result, rows.Err()
}

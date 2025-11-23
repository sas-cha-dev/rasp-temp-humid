package repository

import (
	"database/sql"
	"time"
)

type ButtonReading struct {
	ID       int64     `json:"id"`
	ButtonID int       `json:"button_id"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
}

type ButtonRepository interface {
	Save(buttonID int, startAt time.Time, endAt time.Time) error
	GetLatest() ([]*ButtonReading, error)
}

type buttonRepository struct {
	db *sql.DB
}

func NewButtonRepository(db *sql.DB) (ButtonRepository, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &buttonRepository{db: db}
	if err := repo.createTable(); err != nil {
		return nil, err
	}

	return &buttonRepository{db: db}, nil
}

func (r *buttonRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS button_readings (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    button_id INTEGER NOT NULL,
	    start_at DATETIME NOT NULL,
	    end_at DATETIME NOT NULL
	)`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *buttonRepository) Save(buttonID int, startAt time.Time, endAt time.Time) error {
	query := `INSERT INTO button_readings (button_id, start_at, end_at) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, buttonID, startAt, endAt)
	return err
}

func (r *buttonRepository) GetLatest() ([]*ButtonReading, error) {
	query := `SELECT * FROM button_readings ORDER BY start_at DESC LIMIT 1`
	readings, err := r.queryReadings(query)
	if err != nil {
		return nil, err
	}

	if len(readings) == 0 {
		return nil, nil
	}
	return readings, nil
}

func (r *buttonRepository) queryReadings(query string, args ...interface{}) ([]*ButtonReading, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*ButtonReading
	for rows.Next() {
		var reading ButtonReading
		err := rows.Scan(&reading.ID, &reading.ButtonID, &reading.StartAt, &reading.EndAt)
		if err != nil {
			return nil, err
		}
		readings = append(readings, &reading)
	}

	return readings, rows.Err()
}

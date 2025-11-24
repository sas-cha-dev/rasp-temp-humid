package handler

import (
	"BeRoHuTe/internal/repository"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Repository interface {
	GetLatest() ([]*repository.Reading, error)
	GetLastN(n int) ([]*repository.Reading, error)
	GetAverageLastHour() (map[int]map[string]float64, error)
	GetAverageToday() (map[int]map[string]float64, error)
	GetAverageThisWeek() (map[int]map[string]float64, error)
}

type Handler struct {
	repo        Repository
	btnRepo     repository.ButtonRepository
	indexTpl    *template.Template
	weatherRepo repository.WeatherRepository
}

type DashboardData struct {
	Latest            []*repository.Reading
	LastHour          map[int]map[string]float64
	Today             map[int]map[string]float64
	ThisWeek          map[int]map[string]float64
	Last100           []*repository.Reading
	LastButtonPushes  []*repository.ButtonReading
	LatestWeatherData []*repository.WeatherData
}

func New(repo Repository, templateDir string, btnRepo repository.ButtonRepository,
	weatherRepo repository.WeatherRepository) (*Handler, error) {
	tpl, err := template.ParseFiles(filepath.Join(templateDir, "index.html"))
	if err != nil {
		return nil, err
	}

	return &Handler{
		repo:        repo,
		indexTpl:    tpl,
		btnRepo:     btnRepo,
		weatherRepo: weatherRepo,
	}, nil
}

// ServeIndex renders the main dashboard
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	latest, err := h.repo.GetLatest()
	if err != nil {
		log.Printf("Error getting latest readings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	lastHour, err := h.repo.GetAverageLastHour()
	if err != nil {
		log.Printf("Error getting last hour average: %v", err)
	}

	today, err := h.repo.GetAverageToday()
	if err != nil {
		log.Printf("Error getting today average: %v", err)
	}

	thisWeek, err := h.repo.GetAverageThisWeek()
	if err != nil {
		log.Printf("Error getting this week average: %v", err)
	}

	last100, err := h.repo.GetLastN(100)
	if err != nil {
		log.Printf("Error getting last 100 readings: %v", err)
	}

	lastOpenWindows, err := h.btnRepo.GetLatest()
	if err != nil {
		log.Printf("Error getting last open windows: %v", err)
	}

	lastWeatherData, err := h.weatherRepo.GetLatest()
	if err != nil {
		log.Printf("Error getting last weather data: %v", err)
	}

	data := DashboardData{
		Latest:            latest,
		LastHour:          lastHour,
		Today:             today,
		ThisWeek:          thisWeek,
		Last100:           last100,
		LastButtonPushes:  lastOpenWindows,
		LatestWeatherData: lastWeatherData,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.indexTpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ServeAPI returns JSON data for API requests
func (h *Handler) ServeAPI(w http.ResponseWriter, r *http.Request) {
	latest, err := h.repo.GetLatest()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	lastHour, _ := h.repo.GetAverageLastHour()
	today, _ := h.repo.GetAverageToday()
	thisWeek, _ := h.repo.GetAverageThisWeek()
	last100, _ := h.repo.GetLastN(100)
	lastOpenWindows, _ := h.btnRepo.GetLatest()
	lastWeatherData, _ := h.weatherRepo.GetLatest()

	data := DashboardData{
		Latest:            latest,
		LastHour:          lastHour,
		Today:             today,
		ThisWeek:          thisWeek,
		Last100:           last100,
		LastButtonPushes:  lastOpenWindows,
		LatestWeatherData: lastWeatherData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

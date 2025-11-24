package weather

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	OwExcludeCurrent  string = "current"
	OwExcludeMinutely string = "minutely"
	OwExcludeHourly   string = "hourly"
	OwExcludeDaily    string = "daily"
	OwExcludeAlerts   string = "alerts"
)

type OpenWeatherService struct {
	apiKey    string
	latitude  float64
	longitude float64
	excludes  []string
}

func NewOpenWeatherService(key string, lat, long float64) *OpenWeatherService {
	return &OpenWeatherService{
		apiKey:    key,
		latitude:  lat,
		longitude: long,
		excludes:  []string{OwExcludeHourly, OwExcludeAlerts, OwExcludeDaily, OwExcludeMinutely},
	}
}

func (o *OpenWeatherService) currentWeatherLink() string {
	queryParams := url.Values{}
	queryParams.Set("appid", o.apiKey)
	queryParams.Set("lat", strconv.FormatFloat(o.latitude, 'f', -1, 64))
	queryParams.Set("lon", strconv.FormatFloat(o.longitude, 'f', -1, 64))
	queryParams.Set("exclude", strings.Join(o.excludes, ","))
	queryParams.Set("units", "metric")
	queryParams.Set("lang", "de")

	return "https://api.openweathermap.org/data/3.0/onecall?" + queryParams.Encode()
}

type OpenWeatherOneCallCurrentDetails struct {
	Timestamp   int64   `json:"dt"`
	Temperature float32 `json:"temp"`
	Humidity    float32 `json:"humidity"`
	FeelsLike   float32 `json:"feels_like"`
}

type OpenWeatherOneCallDetails struct {
	Lat      float32                          `json:"lat"`
	Lon      float32                          `json:"lon"`
	Timezone string                           `json:"timezone"`
	Current  OpenWeatherOneCallCurrentDetails `json:"current"`
}

func (o *OpenWeatherService) GetCurrentWeatherDetails() (*CurrentWeather, error) {
	uri := o.currentWeatherLink()

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var owDetails OpenWeatherOneCallDetails
	if err := json.NewDecoder(resp.Body).Decode(&owDetails); err != nil {
		return nil, err
	}

	details := &CurrentWeather{
		Latitude:    owDetails.Lat,
		Longitude:   owDetails.Lon,
		Timestamp:   owDetails.Current.Timestamp,
		Temperature: owDetails.Current.Temperature,
		Humidity:    owDetails.Current.Humidity,
		FeelsLike:   owDetails.Current.FeelsLike,
	}
	log.Println("Fetching current weather details", details)

	return details, nil
}

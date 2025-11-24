package weather

type CurrentWeather struct {
	Latitude    float32 `json:"lat"`
	Longitude   float32 `json:"lon"`
	Timestamp   int64   `json:"timestamp"`
	Temperature float32 `json:"temp"`
	Humidity    float32 `json:"humidity"`
	FeelsLike   float32 `json:"feels_like"`
}

type Service interface {
	GetCurrentWeatherDetails() (*CurrentWeather, error)
}

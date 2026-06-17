package main

import (
	"math"
	"math/rand/v2"
	"time"
)

// TemperatureResponse mirrors the contract expected by the smart_home monolith.
type TemperatureResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

// randomFloat is injectable for deterministic tests; by default returns rand.Float64.
type randomFloat func() float64

// resolveLocation derives a location from a sensor ID and vice versa,
// повторяя поведение, описанное в задании.
func resolveLocation(location, sensorID string) (string, string) {
	// если локация не задана — выбираем по идентификатору сенсора
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// если идентификатор сенсора не задан — выбираем по локации
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}

	return location, sensorID
}

// generateTemperature builds a TemperatureResponse with a random value rounded to one
// decimal place; the value falls in [15.0, 30.0] (rounding may yield exactly 30.0).
func generateTemperature(location, sensorID string, rnd randomFloat, now time.Time) TemperatureResponse {
	location, sensorID = resolveLocation(location, sensorID)

	// случайное значение температуры в диапазоне 15.0..30.0 °C
	const minTemp, maxTemp = 15.0, 30.0
	value := minTemp + rnd()*(maxTemp-minTemp)
	// округляем до одного знака после запятой (math.Round — корректно округляет, не обрезает)
	value = math.Round(value*10) / 10

	return TemperatureResponse{
		Value:       value,
		Unit:        "°C",
		Timestamp:   now,
		Location:    location,
		Status:      "active",
		SensorID:    sensorID,
		SensorType:  "temperature",
		Description: "Temperature reading for " + location,
	}
}

// defaultRandom returns a uniformly distributed float64 in [0.0, 1.0).
func defaultRandom() float64 {
	return rand.Float64()
}

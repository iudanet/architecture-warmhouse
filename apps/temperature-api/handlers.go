package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// server holds dependencies for HTTP handlers (dependency injection for testability).
type server struct {
	rnd randomFloat
	now func() time.Time
}

// newServer creates a server with default dependencies.
func newServer() *server {
	return &server{
		rnd: defaultRandom,
		now: time.Now,
	}
}

// routes registers all HTTP routes on a ServeMux.
func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	// Go 1.22+ ServeMux поддерживает wildcard-сегменты в путях
	mux.HandleFunc("GET /temperature", s.handleTemperatureByLocation)
	mux.HandleFunc("GET /temperature/{sensorID}", s.handleTemperatureByID)
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

// handleHealth responds with a simple liveness check.
func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.methodNotAllowed(w, r)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleTemperatureByLocation handles GET /temperature?location=...
func (s *server) handleTemperatureByLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.methodNotAllowed(w, r)
		return
	}

	location := r.URL.Query().Get("location")
	resp := generateTemperature(location, "", s.rnd, s.now())

	slog.Info("temperature requested by location",
		"location", resp.Location,
		"sensor_id", resp.SensorID,
		"value", resp.Value,
	)

	writeJSON(w, http.StatusOK, resp)
}

// handleTemperatureByID handles GET /temperature/{sensorID}
func (s *server) handleTemperatureByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.methodNotAllowed(w, r)
		return
	}

	// извлекаем идентификатор сенсора из пути /temperature/{sensorID}
	sensorID := r.PathValue("sensorID")
	if sensorID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sensor id is required"})
		slog.Error("missing sensor id", "path", r.URL.Path)
		return
	}

	resp := generateTemperature("", sensorID, s.rnd, s.now())

	slog.Info("temperature requested by sensor id",
		"sensor_id", resp.SensorID,
		"location", resp.Location,
		"value", resp.Value,
	)

	writeJSON(w, http.StatusOK, resp)
}

// methodNotAllowed writes a 405 response and logs the rejected method.
func (s *server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	slog.Error("method not allowed", "method", r.Method, "path", r.URL.Path)
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

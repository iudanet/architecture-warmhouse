package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer creates a server with deterministic dependencies.
func newTestServer(value float64) *server {
	return &server{
		rnd: func() float64 { return value },
		now: func() time.Time { return time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC) },
	}
}

func TestHandleTemperatureByLocation(t *testing.T) {
	srv := newTestServer(0.5)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/temperature?location=Kitchen", nil)

	srv.routes().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp TemperatureResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Kitchen", resp.Location)
	assert.Equal(t, "3", resp.SensorID)
	assert.Equal(t, 22.5, resp.Value)
}

func TestHandleTemperatureByLocation_Empty(t *testing.T) {
	srv := newTestServer(0)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/temperature", nil)

	srv.routes().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp TemperatureResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	// пустая локация -> Unknown (sensorID тоже пуст -> "0")
	assert.Equal(t, "Unknown", resp.Location)
}

func TestHandleTemperatureByID(t *testing.T) {
	srv := newTestServer(0.5)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/temperature/1", nil)

	srv.routes().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp TemperatureResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp.SensorID)
	assert.Equal(t, "Living Room", resp.Location)
	assert.Equal(t, "temperature", resp.SensorType)
}

func TestHandleTemperature_DifferentValues(t *testing.T) {
	// эмулируем меняющийся rnd между запросами
	values := []float64{0.1, 0.8}
	idx := 0
	srv := &server{
		rnd: func() float64 { v := values[idx%len(values)]; idx++; return v },
		now: func() time.Time { return time.Now() },
	}
	mux := srv.routes()

	get := func() float64 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/temperature/1", nil)
		mux.ServeHTTP(rec, req)
		var resp TemperatureResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		return resp.Value
	}

	// два последовательных вызова возвращают разные значения температуры
	assert.NotEqual(t, get(), get())
}

func TestHandleHealth(t *testing.T) {
	srv := newTestServer(0.5)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	srv.routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ok")
}

func TestHandleTemperature_MethodNotAllowed(t *testing.T) {
	srv := newTestServer(0.5)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/temperature?location=Kitchen", nil)

	srv.routes().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

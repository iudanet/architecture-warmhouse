package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLocation(t *testing.T) {
	tests := []struct {
		name         string
		inLocation   string
		inSensorID   string
		wantLocation string
		wantSensorID string
	}{
		{"by sensor id 1", "", "1", "Living Room", "1"},
		{"by sensor id 2", "", "2", "Bedroom", "2"},
		{"by sensor id 3", "", "3", "Kitchen", "3"},
		{"unknown sensor id", "", "99", "Unknown", "99"},
		{"by location living room", "Living Room", "", "Living Room", "1"},
		{"by location bedroom", "Bedroom", "", "Bedroom", "2"},
		{"by location kitchen", "Kitchen", "", "Kitchen", "3"},
		{"unknown location", "Garage", "", "Garage", "0"},
		{"both provided", "Kitchen", "5", "Kitchen", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, sid := resolveLocation(tt.inLocation, tt.inSensorID)
			assert.Equal(t, tt.wantLocation, loc)
			assert.Equal(t, tt.wantSensorID, sid)
		})
	}
}

func TestGenerateTemperature_Range(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)

	// проверяем границы диапазона при rnd=0 и rnd близко к 1
	low := generateTemperature("Living Room", "", func() float64 { return 0 }, now)
	assert.Equal(t, 15.0, low.Value)

	// при rnd≈1 значение приближается к 30.0; после округления до 0.1
	// верхняя граница может стать ровно 30.0 — это допустимо для симулятора
	high := generateTemperature("Living Room", "", func() float64 { return 0.999999 }, now)
	assert.GreaterOrEqual(t, high.Value, 15.0)
	assert.LessOrEqual(t, high.Value, 30.0)
}

func TestGenerateTemperature_Fields(t *testing.T) {
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	resp := generateTemperature("", "2", func() float64 { return 0.5 }, now)

	require.Equal(t, "Bedroom", resp.Location)
	require.Equal(t, "2", resp.SensorID)
	assert.Equal(t, "°C", resp.Unit)
	assert.Equal(t, "active", resp.Status)
	assert.Equal(t, "temperature", resp.SensorType)
	assert.Equal(t, now, resp.Timestamp)
	assert.Contains(t, resp.Description, "Bedroom")
	// при rnd=0.5: 15 + 0.5*15 = 22.5
	assert.Equal(t, 22.5, resp.Value)
}

func TestGenerateTemperature_Randomness(t *testing.T) {
	now := time.Now()
	// последовательные вызовы с разным rnd дают разные значения
	a := generateTemperature("Living Room", "", func() float64 { return 0.1 }, now)
	b := generateTemperature("Living Room", "", func() float64 { return 0.9 }, now)
	assert.NotEqual(t, a.Value, b.Value)
}

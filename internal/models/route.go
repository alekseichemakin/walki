package models

import "time"

type RouteVersion struct {
	ID              int
	RouteID         int
	VersionNumber   int
	Title           string
	Description     string
	DurationMinutes int
	LengthKm        float64
	Theme           string
	Price           float64
	City            string
	CreatedAt       time.Time
	CoverImageURL   string
}

type Route struct {
	ID        int
	Status    string
	IsVisible bool
	CreatedBy int
	CreatedAt time.Time
	UpdatedAt time.Time
	Versions  []RouteVersion // Для полной информации о маршруте
}

package models

import "time"

type RoutePoint struct {
	ID                  int
	VersionID           int
	Title               string
	Description         string
	Latitude            float64
	Longitude           float64
	OrderIndex          int
	Status              string
	ArrivalInstructions string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Media               []Media
}

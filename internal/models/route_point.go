package models

import "time"

type RoutePoint struct {
	ID          int
	VersionID   int
	Idx         int
	Title       string
	Description string
	Lat         float64
	Lon         float64
	CreatedAt   time.Time
}

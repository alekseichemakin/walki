package models

import "time"

type RoutePointLog struct {
	ID         int
	ProgressId int
	PointID    int
	VisitedAt  time.Time
}

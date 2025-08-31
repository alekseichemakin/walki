package models

import "time"

type RouteProgress struct {
	ID           int
	UserID       int
	RouteID      int
	VersionID    int
	CurrentIdx   int
	StartedAt    time.Time
	FinishedAt   *time.Time
	ContentMsgID *int
	VoiceMsgID   *int
}

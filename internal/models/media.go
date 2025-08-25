package models

import "time"

type Media struct {
	ID          int
	Type        string
	URL         string
	Filename    string
	SizeBytes   int64
	Description string
	UploadedBy  int
	UploadedAt  time.Time
	IsPublic    bool
}

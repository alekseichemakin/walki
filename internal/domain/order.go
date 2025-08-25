package domain

import "time"

type OrderSummary struct {
	RouteID      int
	RouteTitle   string
	RouteCity    string
	VersionID    int
	AccessExpiry *time.Time
}

package model

import "time"

type CommissionEvent struct {
	EventID       string
	OrderID       string
	ItemID        string
	ModelID       string
	OrderStatus   string
	OrderedAt     *time.Time
	TrackingTag   string
	Quantity      int
	CommissionTotal int64
}

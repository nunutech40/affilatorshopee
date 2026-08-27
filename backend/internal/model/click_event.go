package model

import "time"

type ClickEvent struct {
	ClickID     string    `json:"click_id"`
	ClickedAt   time.Time `json:"clicked_at"`
	Region      string    `json:"region"`
	TrackingTag string    `json:"tracking_tag"`
	Referrer    string    `json:"referrer"`
}

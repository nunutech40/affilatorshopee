package model

import "time"

type PostLog struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Platform  string    `json:"platform"`
	Caption   string    `json:"caption"`
	Hashtags  []string  `json:"hashtags"`
	Notes     *string   `json:"notes"`
	PostedAt  time.Time `json:"posted_at"`
}

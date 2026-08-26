package model

import "time"

type CaptionVariation struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	Label     string    `json:"label"`
	Template  string    `json:"template"`
	Caption   string    `json:"caption"`
	Hashtags  []string  `json:"hashtags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

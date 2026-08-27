package model

import "time"

type CommissionEvent struct {
	EventID         string
	OrderID         string
	ItemID          string
	ModelID         string
	OrderStatus     string
	OrderedAt       *time.Time
	TrackingTag     string
	Quantity        int
	CommissionTotal int64
}

type SoldProduct struct {
	TrackingTag     string     `json:"tracking_tag"`
	NormalizedTag   string     `json:"normalized_tag"`
	ProductID       *string    `json:"product_id"`
	ProductName     *string    `json:"product_name"`
	ShopeeLink      *string    `json:"shopee_link"`
	ImageURL        *string    `json:"image_url"`
	TotalQuantity   int        `json:"total_quantity"`
	TotalCommission int64      `json:"total_commission"`
	OrderCount      int        `json:"order_count"`
	LastOrderedAt   *time.Time `json:"last_ordered_at"`
	IsInLibrary     bool       `json:"is_in_library"`
}

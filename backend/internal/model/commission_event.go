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
	NormalizedTag   string
	Quantity        int
	CommissionTotal int64
	ItemName        string
	ShopName        string
}

type SoldProduct struct {
	TrackingTag     string     `json:"tracking_tag"`
	NormalizedTag   string     `json:"normalized_tag"`
	ProductID       *string    `json:"product_id"`
	ProductName     *string    `json:"product_name"`
	ShopeeLink      *string    `json:"shopee_link"`
	ImageURL        *string    `json:"image_url"`
	ItemName        *string    `json:"item_name"`
	ItemID          *string    `json:"item_id"`
	ShopName        *string    `json:"shop_name"`
	TotalQuantity   int        `json:"total_quantity"`
	TotalCommission int64      `json:"total_commission"`
	OrderCount      int        `json:"order_count"`
	LastOrderedAt   *time.Time `json:"last_ordered_at"`
	IsInLibrary     bool       `json:"is_in_library"`
}

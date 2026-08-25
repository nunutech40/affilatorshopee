package model

import "time"

type MediaFile struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	SourceURL   string    `json:"source_url"`
	LocalPath   string    `json:"local_path"`
	Filename    string    `json:"filename"`
	MediaType   string    `json:"media_type"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

package model

import "time"

type Product struct {
	ID              string     `json:"id"`
	RawText         string     `json:"raw_text"`
	ProductName     *string    `json:"product_name"`
	ShopeeLink      string     `json:"shopee_link"`
	TrackingTag     string     `json:"tracking_tag"`
	SourceCategory  string     `json:"source_category"`
	ImageURL        *string    `json:"image_url"`
	ImageURLs       []string   `json:"image_urls"`
	VideoURL        *string    `json:"video_url"`
	NormalPrice     *int       `json:"normal_price"`
	SalePrice       *int       `json:"sale_price"`
	DiscountPercent *int       `json:"discount_percent"`
	Rating          *float64   `json:"rating"`
	SoldCount       *string    `json:"sold_count"`
	ReviewCount     *string    `json:"review_count"`
	Keyword         *string    `json:"keyword"`
	Problem         *string    `json:"problem"`
	Cluster         *string    `json:"cluster"`
	ContentModel    *string    `json:"content_model"`
	CaptureAngle    *string    `json:"capture_angle"`
	Benefit1        *string    `json:"benefit_1"`
	Benefit2        *string    `json:"benefit_2"`
	Benefit3        *string    `json:"benefit_3"`
	Urgency         *string    `json:"urgency"`
	CaptionTemplate string     `json:"caption_template"`
	HashtagPool     []string   `json:"hashtag_pool"`
	ReformattedText *string    `json:"reformatted_text"`
	Notes           *string    `json:"notes"`
	Status          string     `json:"status"`
	PostCount       int        `json:"post_count"`
	ClickCount      int        `json:"click_count"`
	LastClickedAt   *time.Time `json:"last_clicked_at"`
	SalesCount      int        `json:"sales_count"`
	PendingSalesCount int      `json:"pending_sales_count"`
	CommissionTotal int64      `json:"commission_total"`
	LastPostedAt    *time.Time `json:"last_posted_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

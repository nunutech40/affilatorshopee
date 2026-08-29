package model

import "time"

type ContentNiche struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ContentStats struct {
	LikeCount     int64     `json:"like_count"`
	RepostCount   int64     `json:"repost_count"`
	ReplyCount    int64     `json:"reply_count"`
	BookmarkCount int64     `json:"bookmark_count"`
	ViewCount     int64     `json:"view_count"`
	CapturedAt    time.Time `json:"captured_at"`
}

type ContentItem struct {
	ID             string           `json:"id"`
	Platform       string           `json:"platform"`
	ExternalPostID string           `json:"external_post_id"`
	CanonicalURL   string           `json:"canonical_url"`
	AuthorHandle   string           `json:"author_handle"`
	OriginalText   string           `json:"original_text"`
	Media          []string         `json:"media"`
	PublishedAt    *time.Time       `json:"published_at"`
	SourceQuery    string           `json:"source_query"`
	Status         string           `json:"status"`
	Niches         []ContentNiche   `json:"niches"`
	ProductTypes   []Niche          `json:"product_types"`
	LatestStats    *ContentStats    `json:"latest_stats,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Variants       []ContentVariant `json:"variants,omitempty"`
}

type ContentVariant struct {
	ID            string    `json:"id"`
	ContentItemID string    `json:"content_item_id"`
	Name          string    `json:"name"`
	Text          string    `json:"text"`
	Source        string    `json:"source"`
	Model         string    `json:"model"`
	Position      int       `json:"position"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

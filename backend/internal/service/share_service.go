package service

import (
	"fmt"
	"net/url"
	"strings"
)

type ShareService struct{}

func NewShareService() *ShareService {
	return &ShareService{}
}

func (s *ShareService) TwitterIntentURL(caption string, media []string) string {
	query := url.Values{}
	query.Set("text", caption)
	for _, item := range media {
		if strings.TrimSpace(item) != "" {
			query.Add("affiliator_media", item)
		}
	}
	return fmt.Sprintf("https://twitter.com/intent/tweet?%s", query.Encode())
}

func (s *ShareService) NormalizeHashtags(hashtags []string) []string {
	normalized := make([]string, 0, len(hashtags))
	seen := map[string]struct{}{}
	for _, tag := range hashtags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, tag)
	}
	return normalized
}

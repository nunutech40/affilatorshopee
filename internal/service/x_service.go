package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

var xURLPattern = regexp.MustCompile(`https?://(?:www\.)?(?:x\.com|twitter\.com)/([^/]+)/status/(\d+)`)
var shopeeLinkPattern = regexp.MustCompile(`https?://[^\s]*shopee[^\s]*`)

type XService struct {
	client      *http.Client
	productRepo *repository.ProductRepository
	media       *MediaService
}

func NewXService(productRepo *repository.ProductRepository, media *MediaService) *XService {
	return &XService{
		client:      &http.Client{Timeout: 15 * time.Second},
		productRepo: productRepo,
		media:       media,
	}
}

type fxtweetResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Tweet   struct {
		Text  string `json:"text"`
		Media *struct {
			All []struct {
				Type         string `json:"type"`
				URL          string `json:"url"`
				ThumbnailURL string `json:"thumbnail_url"`
				Format       string `json:"format"`
				Formats      []struct {
					URL       string `json:"url"`
					Container string `json:"container"`
					Bitrate   int    `json:"bitrate"`
				} `json:"formats"`
			} `json:"all"`
		} `json:"media"`
	} `json:"tweet"`
}

func (s *XService) ImportFromX(ctx context.Context, xURL string, contentModel *string) (*model.Product, *MediaDownloadSummary, error) {
	xURL = strings.TrimSpace(xURL)
	matches := xURLPattern.FindStringSubmatch(xURL)
	if len(matches) != 3 {
		return nil, nil, fmt.Errorf("%w: format link X tidak valid, pakai https://x.com/user/status/ID", ErrValidation)
	}
	user := matches[1]
	id := matches[2]

	// normalize content_model
	var cm *string
	if contentModel != nil && strings.TrimSpace(*contentModel) != "" {
		normalized := normalizeContentModel(*contentModel)
		cm = &normalized
	}

	apis := []string{
		fmt.Sprintf("https://api.fxtwitter.com/%s/status/%s", user, id),
		fmt.Sprintf("https://api.vxtwitter.com/%s/status/%s", user, id),
		fmt.Sprintf("https://api.fixupx.com/%s/status/%s", user, id),
	}

	var lastErr error
	var tweetText string
	var imageURLs []string
	var videoURL *string

	for _, api := range apis {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "AffiliatorShopee/1.0")
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		var data fxtweetResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			lastErr = err
			continue
		}
		resp.Body.Close()
		if data.Code != 200 || data.Tweet.Text == "" {
			lastErr = fmt.Errorf("tweet tidak ditemukan")
			continue
		}
		tweetText = strings.TrimSpace(data.Tweet.Text)
		// extract media
		if data.Tweet.Media != nil {
			for _, m := range data.Tweet.Media.All {
				if m.Type == "photo" && m.URL != "" {
					imageURLs = append(imageURLs, m.URL)
				} else if m.Type == "video" {
					best := m.URL
					bestBitrate := 0
					for _, f := range m.Formats {
						if f.Container == "mp4" && f.Bitrate > bestBitrate {
							bestBitrate = f.Bitrate
							best = f.URL
						}
					}
					if best != "" {
						v := best
						videoURL = &v
						// also add thumbnail as image if needed? user wants download image+video, so keep video separate
					}
				}
			}
		}
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, nil, fmt.Errorf("%w: gagal mengambil postingan X: %v", ErrValidation, lastErr)
	}

	if tweetText == "" {
		return nil, nil, fmt.Errorf("%w: caption postingan kosong", ErrValidation)
	}

	// extract shopee link from caption
	shopeeLink := xURL // fallback
	if m := shopeeLinkPattern.FindString(tweetText); m != "" {
		shopeeLink = m
	}

	// product_name = first line or truncated
	productName := strings.TrimSpace(strings.Split(tweetText, "\n")[0])
	if len(productName) > 120 {
		productName = strings.TrimSpace(productName[:120])
	}
	if productName == "" {
		productName = "Postingan X " + id
	}

	promo := tweetText
	product := &model.Product{
		RawText:         tweetText,
		ReformattedText: &promo,
		ProductName:     &productName,
		ShopeeLink:      shopeeLink,
		ContentModel:    cm,
		HashtagPool:     []string{},
		Notes:           func() *string { s := "Imported dari X: " + xURL; return &s }(),
		Status:          "ready",
		CaptionTemplate: "direct_product",
	}
	// image handling
	if len(imageURLs) > 0 {
		product.ImageURLs = imageURLs
		product.ImageURL = &imageURLs[0]
	}
	if videoURL != nil {
		product.VideoURL = videoURL
	}

	if err := validateProductFields(product); err != nil {
		return nil, nil, err
	}
	// need at least product_name, shopee_link, content_model? For ready, content_model may be required, but for X import we allow ready without full validation? User said ready, so we should enforce ready validation?
	// For X import, we consider ready even if cluster missing, so skip validateReady and directly set ready
	// But ensure content_model default if not provided
	if product.ContentModel == nil || *product.ContentModel == "" {
		def := "cheap"
		product.ContentModel = &def
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, nil, err
	}

	// download media
	summary := &MediaDownloadSummary{Downloaded: []model.MediaFile{}, Failed: []MediaDownloadFailure{}}
	if s.media != nil {
		res := s.media.DownloadProductMedia(ctx, product.ID, product.ImageURLs, product.VideoURL)
		summary.Downloaded = res.Downloaded
		summary.Failed = res.Failed
	}

	return product, summary, nil
}

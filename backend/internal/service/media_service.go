package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
	"github.com/nunutech40/affilatorshopee/internal/storage"
	"golang.org/x/image/webp"
)

const (
	maxImageSize = 20 << 20
	maxVideoSize = 200 << 20
)

type MediaService struct {
	storage storage.Storage
	repo    *repository.MediaRepository
	client  *http.Client
}

type MediaDownloadFailure struct {
	SourceURL string `json:"source_url"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type MediaDownloadSummary struct {
	Downloaded []model.MediaFile      `json:"downloaded"`
	Failed     []MediaDownloadFailure `json:"failed"`
}

func NewMediaService(mediaStorage storage.Storage, repo *repository.MediaRepository) *MediaService {
	client := &http.Client{Timeout: 2 * time.Minute}
	client.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
		return validateExternalURL(req.URL)
	}
	return &MediaService{storage: mediaStorage, repo: repo, client: client}
}

func (s *MediaService) DownloadProductMedia(ctx context.Context, productID string, imageURLs []string, videoURL *string) MediaDownloadSummary {
	result := MediaDownloadSummary{Downloaded: []model.MediaFile{}, Failed: []MediaDownloadFailure{}}
	seen := map[string]bool{}
	imageIndex := 0
	videoIndex := 0
	if existing, err := s.repo.ListByProduct(ctx, productID); err == nil {
		for _, media := range existing {
			seen[strings.TrimSpace(media.SourceURL)] = true
			if media.MediaType == "image" {
				imageIndex++
			}
			if media.MediaType == "video" {
				videoIndex++
			}
		}
	}
	for _, sourceURL := range imageURLs {
		sourceURL = strings.TrimSpace(sourceURL)
		if sourceURL == "" || seen[sourceURL] {
			continue
		}
		seen[sourceURL] = true
		imageIndex++
		media, err := s.download(ctx, productID, sourceURL, "image", imageIndex)
		if err != nil {
			result.Failed = append(result.Failed, MediaDownloadFailure{SourceURL: sourceURL, Code: "MEDIA_DOWNLOAD_FAILED", Message: err.Error()})
			continue
		}
		result.Downloaded = append(result.Downloaded, *media)
	}
	if videoURL != nil {
		sourceURL := strings.TrimSpace(*videoURL)
		if sourceURL != "" && !seen[sourceURL] {
			seen[sourceURL] = true
			videoIndex++
			media, err := s.download(ctx, productID, sourceURL, "video", videoIndex)
			if err != nil {
				result.Failed = append(result.Failed, MediaDownloadFailure{SourceURL: sourceURL, Code: "MEDIA_DOWNLOAD_FAILED", Message: err.Error()})
			} else {
				result.Downloaded = append(result.Downloaded, *media)
			}
		}
	}
	return result
}

func (s *MediaService) List(ctx context.Context, productID string) ([]model.MediaFile, error) {
	return s.repo.ListByProduct(ctx, productID)
}

func (s *MediaService) Open(ctx context.Context, productID, mediaID string) (*model.MediaFile, *os.File, error) {
	item, err := s.repo.GetByID(ctx, productID, mediaID)
	if err != nil {
		return nil, nil, err
	}
	file, err := s.storage.Open(item.LocalPath)
	if err != nil {
		return nil, nil, err
	}
	return item, file, nil
}

func (s *MediaService) Remove(ctx context.Context, productID, mediaID string) error {
	item, err := s.repo.GetByID(ctx, productID, mediaID)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, productID, mediaID); err != nil {
		return err
	}
	if err := s.storage.Delete(item.LocalPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *MediaService) Zip(ctx context.Context, productID string) (*bytes.Buffer, error) {
	items, err := s.List(ctx, productID)
	if err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)
	for _, item := range items {
		file, err := s.storage.Open(item.LocalPath)
		if err != nil {
			_ = archive.Close()
			return nil, err
		}
		entry, err := archive.Create(item.Filename)
		if err == nil {
			_, err = io.Copy(entry, file)
		}
		_ = file.Close()
		if err != nil {
			_ = archive.Close()
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buffer, nil
}

func (s *MediaService) download(ctx context.Context, productID, sourceURL, expectedType string, index int) (*model.MediaFile, error) {
	parsed, err := url.ParseRequestURI(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("URL tidak valid")
	}
	if err := validateExternalURL(parsed); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "AffiliatorShopee/1.0")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("server media mengembalikan status %d", response.StatusCode)
	}
	maxSize := int64(maxImageSize)
	if expectedType == "video" {
		maxSize = maxVideoSize
	}
	if response.ContentLength > maxSize {
		return nil, fmt.Errorf("file melebihi batas %d MB", maxSize>>20)
	}
	header := make([]byte, 512)
	read, err := io.ReadFull(response.Body, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	header = header[:read]
	contentType := response.Header.Get("Content-Type")
	if semi := strings.IndexByte(contentType, ';'); semi >= 0 {
		contentType = contentType[:semi]
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(header)
	}
	if !allowedMediaType(expectedType, contentType, filepath.Ext(parsed.Path)) {
		return nil, fmt.Errorf("content type %s tidak sesuai media %s", contentType, expectedType)
	}
	extension := mediaExtension(contentType, filepath.Ext(parsed.Path), expectedType)
	filename := fmt.Sprintf("%s-%02d%s", expectedType, index, extension)
	reader := io.Reader(io.MultiReader(bytes.NewReader(header), response.Body))
	if expectedType == "image" && contentType == "image/webp" {
		raw, readErr := io.ReadAll(io.LimitReader(reader, maxSize+1))
		if readErr != nil {
			return nil, readErr
		}
		decoded, decodeErr := webp.Decode(bytes.NewReader(raw))
		if decodeErr != nil {
			return nil, fmt.Errorf("decode WebP: %w", decodeErr)
		}
		converted := &bytes.Buffer{}
		if encodeErr := jpeg.Encode(converted, decoded, &jpeg.Options{Quality: 92}); encodeErr != nil {
			return nil, fmt.Errorf("encode JPEG: %w", encodeErr)
		}
		reader = converted
		contentType = "image/jpeg"
		filename = fmt.Sprintf("%s-%02d.jpg", expectedType, index)
	}
	localPath, size, err := s.storage.Save(productID, filename, reader, maxSize)
	if err != nil {
		return nil, err
	}
	media := &model.MediaFile{ProductID: productID, SourceURL: sourceURL, LocalPath: localPath, Filename: filename, MediaType: expectedType, ContentType: contentType, SizeBytes: size}
	if err := s.repo.Create(ctx, media); err != nil {
		_ = s.storage.Delete(localPath)
		return nil, err
	}
	return media, nil
}

func validateExternalURL(parsed *url.URL) error {
	if parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
		return fmt.Errorf("media URL harus http/https tanpa userinfo")
	}
	addresses, err := net.LookupIP(parsed.Hostname())
	if err != nil {
		return fmt.Errorf("host media tidak dapat di-resolve")
	}
	for _, address := range addresses {
		if address.IsLoopback() || address.IsPrivate() || address.IsLinkLocalUnicast() || address.IsUnspecified() {
			return fmt.Errorf("media URL mengarah ke jaringan private")
		}
	}
	return nil
}

func allowedMediaType(expected, contentType, extension string) bool {
	if expected == "video" {
		return strings.HasPrefix(contentType, "video/") || strings.EqualFold(extension, ".mp4") || strings.EqualFold(extension, ".mov") || strings.EqualFold(extension, ".webm")
	}
	return strings.HasPrefix(contentType, "image/") || strings.Contains(".jpg .jpeg .png .gif .webp", strings.ToLower(extension))
}

func mediaExtension(contentType, sourceExtension, expected string) string {
	extension := strings.ToLower(filepath.Ext(sourceExtension))
	if expected == "video" && (extension == ".mp4" || extension == ".mov" || extension == ".webm") {
		return extension
	}
	if expected == "image" && (extension == ".jpg" || extension == ".jpeg" || extension == ".png" || extension == ".gif" || extension == ".webp") {
		return extension
	}
	for suffix, mimeType := range map[string]string{".jpg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".webp": "image/webp", ".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime"} {
		if contentType == mimeType {
			return suffix
		}
	}
	if expected == "video" {
		return ".mp4"
	}
	return ".jpg"
}

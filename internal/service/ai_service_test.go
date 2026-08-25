package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

func TestValidateAIResultsRejectsInvalidModelAndPrice(t *testing.T) {
	name := "Produk"
	products := []model.Product{{ID: "12345678-1234-1234-1234-123456789012", RawText: "raw", ProductName: &name, ShopeeLink: "https://shopee.co.id/x", Status: "raw"}}
	normal, sale := 10, 20
	results := []AIReformatResult{{ProductID: products[0].ID, NormalPrice: &normal, SalePrice: &sale, ContentModel: stringPtr("branded")}}
	if err := validateAIResults(products, results); err == nil {
		t.Fatal("expected invalid AI result")
	}
}

func TestAIServiceReformatParsesProviderResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"[{\"product_id\":\"12345678-1234-1234-1234-123456789012\",\"content_model\":\"cheap\",\"hashtag_pool\":[\"#deal\"]}]"}}]}`))
	}))
	defer server.Close()

	service := NewAIService("test-key", "test-model")
	service.endpoint = server.URL
	service.client = server.Client()
	results, err := service.Reformat(context.Background(), []model.Product{{ID: "12345678-1234-1234-1234-123456789012", RawText: "raw", ShopeeLink: "https://shopee.co.id/x"}}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ContentModel == nil || *results[0].ContentModel != "cheap" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestAIServiceTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(20 * time.Millisecond) }))
	defer server.Close()
	service := NewAIService("test-key", "test-model")
	service.endpoint = server.URL
	service.client = &http.Client{Timeout: 1 * time.Millisecond}
	_, err := service.Reformat(context.Background(), []model.Product{{ID: "12345678-1234-1234-1234-123456789012", RawText: "raw", ShopeeLink: "https://shopee.co.id/x"}}, "")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func stringPtr(value string) *string { return &value }

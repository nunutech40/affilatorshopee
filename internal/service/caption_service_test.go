package service

import (
	"strings"
	"testing"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

func TestCaptionServiceGenerateIncludesHashtagsAndOmitsEmptyLines(t *testing.T) {
	name := "Rak serbaguna"
	benefit := "Hemat tempat"
	product := &model.Product{Status: "ready", ProductName: &name, ShopeeLink: "https://shopee.co.id/item", Benefit1: &benefit}
	result, err := NewCaptionService(NewShareService()).Generate(product, "direct_product", []string{"rumah", "#Rumah"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Caption, "#rumah") || strings.Count(strings.ToLower(result.Caption), "#rumah") != 1 {
		t.Fatalf("hashtags missing: %q", result.Caption)
	}
	if strings.Contains(result.Caption, "✅  terjual") || strings.Contains(result.Caption, "{benefit_2}") {
		t.Fatalf("empty placeholders leaked: %q", result.Caption)
	}
	if result.CharacterCount != RuneLen(result.Caption) {
		t.Fatalf("character count mismatch")
	}
}

func TestCaptionServiceProblemTemplateRequiresProblem(t *testing.T) {
	name := "Lampu meja"
	product := &model.Product{Status: "ready", ProductName: &name, ShopeeLink: "https://shopee.co.id/item"}
	_, err := NewCaptionService(NewShareService()).Generate(product, "problem_specific", nil)
	if err == nil || !strings.Contains(err.Error(), "problem wajib") {
		t.Fatalf("expected problem validation, got %v", err)
	}
}

func TestCaptionServiceSupportsKeywordAndCheapTemplates(t *testing.T) {
	name := "Organizer meja"
	keyword := "organizer meja minimalis"
	benefit := "Merapikan barang"
	price := 29999
	product := &model.Product{Status: "ready", ProductName: &name, Keyword: &keyword, ShopeeLink: "https://shopee.co.id/item", Benefit1: &benefit, SalePrice: &price}
	service := NewCaptionService(NewShareService())
	for _, template := range []string{"keyword_recommendation", "cheap_value"} {
		result, err := service.Generate(product, template, []string{"#meja"})
		if err != nil {
			t.Fatalf("template %s: %v", template, err)
		}
		if !strings.Contains(result.Caption, "#meja") {
			t.Fatalf("template %s omitted hashtag", template)
		}
	}
}

func TestFormatPrice(t *testing.T) {
	value := 39999
	if got := formatPrice(&value); got != "Rp39.999" {
		t.Fatalf("got %q", got)
	}
}

func TestCaptionServiceReportsOverLimit(t *testing.T) {
	name := strings.Repeat("produk ", 50)
	product := &model.Product{Status: "ready", ProductName: &name, ShopeeLink: "https://shopee.co.id/item"}
	result, err := NewCaptionService(NewShareService()).Generate(product, "direct_product", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OverLimit || result.CharacterCount <= 280 {
		t.Fatalf("expected over-limit caption, got %d", result.CharacterCount)
	}
}

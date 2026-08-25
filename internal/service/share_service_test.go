package service

import (
	"net/url"
	"strings"
	"testing"
)

func TestTwitterIntentURLEncodesUnicodeAndNewlines(t *testing.T) {
	caption := "Halo dunia 👋\nCek: https://shopee.co.id/x"
	result := NewShareService().TwitterIntentURL(caption)
	value, err := url.Parse(result)
	if err != nil {
		t.Fatal(err)
	}
	if value.Query().Get("text") != caption {
		t.Fatalf("decoded caption mismatch: %q", value.Query().Get("text"))
	}
	if strings.Contains(result, "Halo dunia") {
		t.Fatalf("caption was not encoded: %s", result)
	}
}

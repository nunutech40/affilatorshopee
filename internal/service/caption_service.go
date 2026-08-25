package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type GeneratedCaption struct {
	Caption        string   `json:"caption"`
	CharacterCount int      `json:"character_count"`
	OverLimit      bool     `json:"over_limit"`
	Template       string   `json:"template"`
	Hashtags       []string `json:"hashtags"`
}

type CaptionService struct {
	share *ShareService
}

func NewCaptionService(share *ShareService) *CaptionService {
	return &CaptionService{share: share}
}

func (s *CaptionService) Generate(product *model.Product, template string, hashtags []string) (GeneratedCaption, error) {
	if product == nil {
		return GeneratedCaption{}, fmt.Errorf("%w: product tidak ditemukan", ErrNotFound)
	}
	if product.ProductName == nil || strings.TrimSpace(*product.ProductName) == "" {
		return GeneratedCaption{}, fmt.Errorf("%w: product belum memiliki product_name", ErrValidation)
	}
	if product.Status != "reformatted" && product.Status != "ready" {
		return GeneratedCaption{}, fmt.Errorf("%w: product belum siap membuat caption", ErrValidation)
	}
	if template == "" {
		template = product.CaptionTemplate
	}
	if !isTemplate(template) {
		return GeneratedCaption{}, fmt.Errorf("%w: template tidak dikenal", ErrValidation)
	}

	normalized := s.share.NormalizeHashtags(hashtags)
	if len(normalized) > 3 {
		return GeneratedCaption{}, fmt.Errorf("%w: maksimal 3 hashtag", ErrValidation)
	}

	values := captionValues(product, normalized)
	text, err := renderTemplate(template, values)
	if err != nil {
		return GeneratedCaption{}, err
	}

	return GeneratedCaption{
		Caption:        text,
		CharacterCount: RuneLen(text),
		OverLimit:      RuneLen(text) > 280,
		Template:       template,
		Hashtags:       normalized,
	}, nil
}

func (s *CaptionService) GenerateVariations(product *model.Product, template string, count int, hashtags []string) ([]GeneratedCaption, error) {
	if count < 2 || count > 3 {
		return nil, fmt.Errorf("%w: jumlah variasi harus 2-3", ErrValidation)
	}

	variations := make([]GeneratedCaption, 0, count)
	templates := []string{template, "keyword_recommendation", "cheap_value"}
	for i := 0; i < count; i++ {
		selected := templates[i]
		if i == 0 && selected == "" {
			selected = product.CaptionTemplate
		}
		generated, err := s.Generate(product, selected, hashtags)
		if err != nil {
			return nil, err
		}
		variations = append(variations, generated)
	}
	return variations, nil
}

func isTemplate(template string) bool {
	switch template {
	case "direct_product", "keyword_recommendation", "problem_specific", "cheap_value":
		return true
	default:
		return false
	}
}

func captionValues(product *model.Product, hashtags []string) map[string]string {
	value := func(input *string) string {
		if input == nil {
			return ""
		}
		return strings.TrimSpace(*input)
	}
	integer := func(input *int) string {
		if input == nil {
			return ""
		}
		return strconv.Itoa(*input)
	}
	float := func(input *float64) string {
		if input == nil {
			return ""
		}
		return strconv.FormatFloat(*input, 'f', 1, 64)
	}

	keyword := value(product.Keyword)
	if keyword == "" {
		keyword = value(product.ProductName)
	}
	return map[string]string{
		"product_name":     value(product.ProductName),
		"keyword":          keyword,
		"problem":          value(product.Problem),
		"benefit_1":        value(product.Benefit1),
		"benefit_2":        value(product.Benefit2),
		"benefit_3":        value(product.Benefit3),
		"rating":           float(product.Rating),
		"sold_count":       value(product.SoldCount),
		"review_count":     value(product.ReviewCount),
		"normal_price":     formatPrice(product.NormalPrice),
		"sale_price":       formatPrice(product.SalePrice),
		"discount_percent": integer(product.DiscountPercent),
		"urgency":          value(product.Urgency),
		"shopee_link":      product.ShopeeLink,
		"hashtags":         strings.Join(hashtags, " "),
	}
}

func renderTemplate(template string, values map[string]string) (string, error) {
	var lines []string
	switch template {
	case "direct_product":
		lines = []string{
			"Cari {product_name}?", "", "✅ {benefit_1}", "✅ {benefit_2}", "✅ {benefit_3}",
			"✅ {rating}⭐️", "✅ {sold_count} terjual", "{urgency}", "", "Cek di sini 👇", "{shopee_link}", "", "{hashtags}",
		}
	case "keyword_recommendation":
		lines = []string{
			"Lagi cari {keyword}?", "", "Ini salah satu yang gue shortlist.", "", "Kenapa masuk shortlist:",
			"✅ {benefit_1}", "✅ {benefit_2}", "✅ {benefit_3}", "✅ {rating}⭐️ | {sold_count} terjual",
			"", "Harganya masih masuk akal.", "", "Cek:", "{shopee_link}", "", "{hashtags}",
		}
	case "problem_specific":
		if values["problem"] == "" {
			return "", fmt.Errorf("%w: problem wajib diisi untuk template problem_specific", ErrValidation)
		}
		lines = []string{
			"Punya masalah {problem}?", "", "{product_name} ini bisa jadi salah satu opsi.", "",
			"✅ {benefit_1}", "✅ {benefit_2}", "✅ {benefit_3}", "", "Cek detailnya:", "{shopee_link}", "", "{hashtags}",
		}
	case "cheap_value":
		lines = []string{
			"Cari {product_name} murah tapi nggak murahan?", "", "Yang ini menarik 👀", "", "✅ {benefit_1}",
			"✅ {benefit_2}", "✅ {rating}⭐️", "✅ {sold_count} terjual", "✅ {sale_price}", "",
			"Kalau budget lo sekitar {sale_price}, ini worth checking.", "", "👇", "{shopee_link}", "", "{hashtags}",
		}
	}

	for i, line := range lines {
		for key, value := range values {
			if value == "" && strings.HasPrefix(strings.TrimSpace(line), "✅") && strings.Contains(line, "{"+key+"}") {
				line = ""
				break
			}
			line = strings.ReplaceAll(line, "{"+key+"}", value)
		}
		lines[i] = strings.TrimSpace(line)
	}

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "✅ ") && strings.TrimSpace(strings.TrimPrefix(line, "✅ ")) == "" {
			continue
		}
		if strings.HasPrefix(line, "✅ ") && strings.TrimSpace(strings.TrimPrefix(line, "✅ ")) == "⭐️" {
			continue
		}
		if strings.TrimSpace(line) == "" {
			if len(result) == 0 || result[len(result)-1] == "" {
				continue
			}
		}
		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n")), nil
}

func formatPrice(value *int) string {
	if value == nil {
		return ""
	}
	n := *value
	if n == 0 {
		return "Rp0"
	}
	s := strconv.Itoa(n)
	start := 0
	if s[0] == '-' {
		start = 1
	}
	for i := len(s) - 3; i > start; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "Rp" + s
}

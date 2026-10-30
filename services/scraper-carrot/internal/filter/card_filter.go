package filter

import (
	"regexp"
	"strconv"
	"strings"

	"scraper-carrot/internal/models"
)

var bulkRegex = regexp.MustCompile(`[2-9]\d+장|1\d{2,}장`)

// Placeholder prices sellers use instead of a real one.
var fakePrices = []float64{123456, 12345}

var stopWords = []string{
	"피규어", "인형", "띠부씰", "케이스", "빈박스", "빈 박스", "닌텐도", "칩", "가챠", "figure",
	"일괄", "처분", "대량", "모음", "다량", "랜덤", "통째로", "정리",
	"각종", "등등", "개별", "각각",
}

type CardFilter struct{}

func NewCardFilter() *CardFilter {
	return &CardFilter{}
}

func (f *CardFilter) IsValid(item models.Article) bool {
	priceFloat, _ := strconv.ParseFloat(item.Price, 64)
	if priceFloat < 3000 {
		return false
	}

	for _, fake := range fakePrices {
		if priceFloat == fake {
			return false
		}
	}

	lowerTitle := strings.ToLower(item.Title)
	lowerContent := strings.ToLower(item.Content)

	for _, word := range stopWords {
		if strings.Contains(lowerTitle, word) || strings.Contains(lowerContent, word) {
			return false
		}
	}

	if bulkRegex.MatchString(lowerTitle) || bulkRegex.MatchString(lowerContent) {
		return false
	}

	return true
}

// Apply returns the articles that pass IsValid.
func (f *CardFilter) Apply(articles []models.Article) []models.Article {
	valid := make([]models.Article, 0, len(articles))
	for _, a := range articles {
		if f.IsValid(a) {
			valid = append(valid, a)
		}
	}
	return valid
}

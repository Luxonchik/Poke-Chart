package filter

import (
	"regexp"
	"strconv"
	"strings"

	"scraper-carrot/internal/models"
)

var bulkRegex = regexp.MustCompile(`[2-9]\d+장|1\d{2,}장`)

var fakePrices = []string{"123456.0", "12345.0"}

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
		if item.Price == fake {
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

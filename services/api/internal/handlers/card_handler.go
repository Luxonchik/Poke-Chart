package handlers

import (
	"encoding/json"
	"net/http"
	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
	"strconv"
)

type CardHandler struct {
	repo *repository.CardRepo
}

func NewCardHandler(repo *repository.CardRepo) *CardHandler {
	return &CardHandler{repo: repo}
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page := 1

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 30
	offset := (page - 1) * limit

	cards, err := h.repo.GetAllCardsRankedByListings(limit, offset)
	h.writeCards(w, cards, err)
}

const maxSearchLen = 100

// GetCatalog — GET /api/catalog?page=N&q=...&on_sale=1 : all master cards, 30 per page.
func (h *CardHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := 1
	if p, err := strconv.Atoi(query.Get("page")); err == nil && p > 0 {
		page = p
	}
	search := query.Get("q")
	if len([]rune(search)) > maxSearchLen {
		search = string([]rune(search)[:maxSearchLen])
	}
	onSaleOnly := query.Get("on_sale") == "1"
	limit := 30
	cards, err := h.repo.GetCatalog(search, onSaleOnly, limit, (page-1)*limit)
	h.writeCards(w, cards, err)
}

func (h *CardHandler) writeCards(w http.ResponseWriter, cards []models.CardWithListingCount, err error) {
	if err != nil {
		http.Error(w, "Ошибка сервера при получении карт", http.StatusInternalServerError)
		return
	}

	if cards == nil {
		cards = []models.CardWithListingCount{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}
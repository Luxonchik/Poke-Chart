package handlers

import (
	"encoding/json"
	"net/http"
	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
	"strconv"
)

// CardHandler хранит в себе ссылку на репозиторий, чтобы уметь ходить в базу
type CardHandler struct {
	repo *repository.CardRepo
}

func NewCardHandler(repo *repository.CardRepo) *CardHandler {
	return &CardHandler{repo: repo}
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	// 1. Просим репозиторий достать карточки. Пока поставим лимит 50 для теста.
	pageStr := r.URL.Query().Get("page")
	page := 1

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	
	limit := 30
	offset := (page - 1) * limit	

	cards, err := h.repo.GetAllCards(limit, offset)
	if err != nil {
		http.Error(w, "Ошибка сервера при получении карт", http.StatusInternalServerError)
		return
	}

	// Защита от null, если карты кончились (Go может вернуть nil-слайс, а фронт ждет пустой массив [])
	if cards == nil {
		cards = []models.Card{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}
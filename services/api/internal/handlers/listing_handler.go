package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
)

type ListingHandler struct {
	repo *repository.ListingRepo
}

func NewListingHandler(repo *repository.ListingRepo) *ListingHandler {
	return &ListingHandler{repo: repo}
}

// ReceiveFromScraper принимает JSON от твоего парсера 당근마켓
func (h *ListingHandler) ReceiveFromScraper(w http.ResponseWriter, r *http.Request) {
	var reqBody models.CreateListingRequest

	// Читаем JSON от скрапера
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Ошибка чтения JSON", http.StatusBadRequest)
		return
	}

	// Если скрапер прислал пустой ID, ругаемся
	if reqBody.DaangnID == "" {
		http.Error(w, "Отсутствует daangn_id", http.StatusBadRequest)
		return
	}

	// Пытаемся сохранить в БД
	isNew, err := h.repo.CreateListing(reqBody)
	if err != nil {
		http.Error(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отвечаем скраперу
	w.Header().Set("Content-Type", "application/json")
	if isNew {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"status": "saved", "message": "Новое объявление добавлено в очередь"}`)
	} else {
		// 200 OK, но говорим, что это дубликат
		w.WriteHeader(http.StatusOK) 
		fmt.Fprintf(w, `{"status": "ignored", "message": "Объявление уже существует"}`)
	}
}
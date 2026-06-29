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

func (h *ListingHandler) GetListingsForCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("cardId")
	if cardID == "" {
		http.Error(w, "Отсутствует cardId", http.StatusBadRequest)
		return
	}

	listings, err := h.repo.GetListingsByMasterCardID(cardID)
	if err != nil {
		http.Error(w, "Ошибка при получении листингов", http.StatusInternalServerError)
		return
	}

	if listings == nil {
		listings = []models.MarketListing{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

func (h *ListingHandler) ReceiveFromScraper(w http.ResponseWriter, r *http.Request) {
	var reqBody models.CreateListingRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Ошибка чтения JSON", http.StatusBadRequest)
		return
	}

	if reqBody.DaangnID == "" {
		http.Error(w, "Отсутствует daangn_id", http.StatusBadRequest)
		return
	}

	isNew, err := h.repo.CreateListing(reqBody)
	if err != nil {
		http.Error(w, "Ошибка базы данных: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if isNew {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"status": "saved", "message": "Новое объявление добавлено в очередь"}`)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ignored", "message": "Объявление уже существует"}`)
	}
}
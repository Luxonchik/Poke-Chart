package handlers

import (
	"encoding/json"
	"log"
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
		log.Printf("❌ GetListingsByMasterCardID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if listings == nil {
		listings = []models.MarketListing{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listings); err != nil {
		log.Printf("❌ JSON encode: %v", err)
	}
}

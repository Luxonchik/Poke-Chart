package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"pokechart-api/internal/repository"
)

type StatsHandler struct {
	repo *repository.ListingRepo
}

func NewStatsHandler(repo *repository.ListingRepo) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetMarketStats()
	if err != nil {
		log.Printf("❌ GetMarketStats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("❌ JSON encode: %v", err)
	}
}

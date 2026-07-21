package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
)

type SyncHandler struct {
	repo *repository.ListingRepo
}

func NewSyncHandler(repo *repository.ListingRepo) *SyncHandler {
	return &SyncHandler{repo: repo}
}

func (h *SyncHandler) Sync(w http.ResponseWriter, r *http.Request) {
	var req models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ожидается JSON {source, listings}", http.StatusBadRequest)
		return
	}

	if req.Source == "" {
		http.Error(w, "Отсутствует source", http.StatusBadRequest)
		return
	}

	result, err := h.repo.SyncListings(req.Source, req.Listings)
	if err != nil {
		log.Printf("❌ SyncListings source=%s: %v", req.Source, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Sync source=%s new=%d updated=%d closed=%d",
		req.Source, result.New, result.Updated, result.Closed)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("❌ JSON encode: %v", err)
	}
}

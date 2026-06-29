package handlers

import (
	"encoding/json"
	"net/http"

	"pokechart-api/internal/services/vision"
)

// TestVisionHandler — временный эндпоинт для ручного тестирования Gemini.
func TestVisionHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		ImageURL string `json:"image_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Ожидается JSON с полем image_url", http.StatusBadRequest)
		return
	}

	result, err := vision.RecognizeCard(reqBody.ImageURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
package handlers

import (
	"encoding/json"
	"net/http"

	"pokechart-api/internal/services/vision"
)

// TestVisionHandler - временный обработчик для тестирования API
func TestVisionHandler(w http.ResponseWriter, r *http.Request) {
	// Принимаем JSON с url картинки
	var reqBody struct {
		ImageURL string `json:"image_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Ожидается JSON с полем image_url", http.StatusBadRequest)
		return
	}

	// Вызываем наш сервис
	result, err := vision.RecognizeCard(reqBody.ImageURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
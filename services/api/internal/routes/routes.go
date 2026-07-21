package routes

import (
	"net/http"

	"pokechart-api/internal/handlers"
	"pokechart-api/internal/repository"
)

func Setup(mux *http.ServeMux, cardRepo *repository.CardRepo, listingRepo *repository.ListingRepo) {
	cardHandler := handlers.NewCardHandler(cardRepo)
	listingHandler := handlers.NewListingHandler(listingRepo)
	syncHandler := handlers.NewSyncHandler(listingRepo)

	mux.HandleFunc("GET /api/cards", cardHandler.GetCards)
	mux.HandleFunc("GET /api/cards/{cardId}/listings", listingHandler.GetListingsForCard)
	mux.HandleFunc("POST /api/sync", syncHandler.Sync)
}

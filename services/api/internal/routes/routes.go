package routes

import (
    "net/http"
    "pokechart-api/internal/handlers"
    "pokechart-api/internal/repository"
    "database/sql"
)

func Setup(mux *http.ServeMux, db *sql.DB) {
    cardRepo := repository.NewCardRepo(db)
    cardHandler := handlers.NewCardHandler(cardRepo)

    listingRepo := repository.NewListingRepo(db)
    listingHandler := handlers.NewListingHandler(listingRepo)

    mux.HandleFunc("GET /api/cards", cardHandler.GetCards)
    mux.HandleFunc("GET /api/cards/{cardId}/listings", listingHandler.GetListingsForCard)
    mux.HandleFunc("POST /api/test-vision", handlers.TestVisionHandler)
    mux.HandleFunc("POST /api/listings", listingHandler.ReceiveFromScraper)
}
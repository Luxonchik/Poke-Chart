package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pokechart-api/internal/repository"
	"pokechart-api/internal/routes"
	"pokechart-api/internal/workers"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ .env не найден, используем переменные окружения")
	}

	db := mustConnectDB()
	defer db.Close()

	cardRepo := repository.NewCardRepo(db)
	listingRepo := repository.NewListingRepo(db)

	mux := http.NewServeMux()
	routes.Setup(mux, cardRepo, listingRepo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	allWorkers := []workers.Worker{
		workers.NewVisionWorker(db, cardRepo),
	}
	for _, w := range allWorkers {
		go w.Start(ctx)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: withCORS(mux),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	fmt.Println("🚀 API запущен на :8080")
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	fmt.Println("👋 API остановлен")
}

func mustConnectDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Ошибка настройки БД:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("❌ БД недоступна:", err)
	}
	fmt.Println("✅ Подключение к БД установлено")
	return db
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

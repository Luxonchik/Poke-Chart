package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"pokechart-api/internal/repository"
	"pokechart-api/internal/routes"
	"pokechart-api/internal/workers"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func enableCORS(next http.Handler) http.Handler {
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

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ Файл .env не найден, надеемся, что переменные среды заданы в системе")
	}

	connStr := "postgres://admin:admin_password@localhost:5433/pokecart?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Ошибка настройки БД:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("❌ БД недоступна:", err)
	}
	fmt.Println("✅ Успешное подключение к базе данных!")

	cardRepo := repository.NewCardRepo(db)
	go workers.StartVisionWorker(db, cardRepo)
	fmt.Println("🤖 AI-воркер запущен в фоне")


	mux := http.NewServeMux()

	routes.Setup(mux, db)

	port := ":8080"
	fmt.Printf("🚀 API сервер запущен и ждет запросы на http://localhost%s\n", port)
	
	handlerWithCORS := enableCORS(mux)

	if err := http.ListenAndServe(port, handlerWithCORS); err != nil {
		log.Fatal("❌ Ошибка запуска сервера:", err)
	}
}
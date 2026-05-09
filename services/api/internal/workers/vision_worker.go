package workers

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"pokechart-api/internal/repository"
	"pokechart-api/internal/services/vision"
)

// StartVisionWorker запускает бесконечный цикл обработки очереди
func StartVisionWorker(db *sql.DB, cardRepo *repository.CardRepo) {
	fmt.Println("🤖 Фоновый AI-воркер запущен и ждет задачи...")

	for {
		var id int
		var imageUrl string

		// 1. Берем ОДНУ задачу из очереди (магия SKIP LOCKED)
		query := `
			UPDATE market_listings 
			SET status = 'processing', updated_at = NOW()
			WHERE id = (
				SELECT id FROM market_listings 
				WHERE status = 'pending' 
				ORDER BY updated_at ASC, created_at ASC
				FOR UPDATE SKIP LOCKED 
				LIMIT 1
			) 
			RETURNING id, image_url;
		`

		err := db.QueryRow(query).Scan(&id, &imageUrl)

		if err == sql.ErrNoRows {
			// Очередь пуста. Спим 10 секунд и проверяем снова.
			time.Sleep(10 * time.Second)
			continue
		} else if err != nil {
			log.Printf("⚠️ Ошибка БД при чтении очереди: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Printf("🔍 Воркер взял в работу объявление ID: %d...\n", id)

		// 2. Отправляем фото в Gemini
		aiResult, err := vision.RecognizeCard(imageUrl)

		if err != nil {
			// Возвращаем задачу в pending, но не блокируем всю очередь одной "проблемной" карточкой.
			log.Printf("❌ Ошибка ИИ: %v. Откладываем задачу...", err)
			db.Exec("UPDATE market_listings SET status = 'pending', ai_error_msg = $1, updated_at = NOW() WHERE id = $2", err.Error(), id)
			time.Sleep(5 * time.Second)
			continue
		}

		// Защита от фото кроссовок или мебели
		if !aiResult.IsPokemonCard {
			db.Exec("UPDATE market_listings SET status = 'rejected', ai_error_msg = 'Не покемон', updated_at = NOW() WHERE id = $1", id)
			continue
		}

		// 3. Ищем карту в официальном справочнике
		masterCard, err := cardRepo.FindCardByNameAndRarity(aiResult.Name, aiResult.Rarity)
		if err != nil {
			// ИИ распознал, но такой карты нет в нашей базе
			log.Printf("⚠️ Карта %s %s не найдена в БД\n", aiResult.Name, aiResult.Rarity)
			db.Exec("UPDATE market_listings SET status = 'error', ai_error_msg = 'Нет в справочнике', updated_at = NOW() WHERE id = $1", id)
			continue
		}

		// 4. Успех! Привязываем объявление к карте и делаем его активным
		_, err = db.Exec(`
			UPDATE market_listings 
			SET status = 'active', master_card_id = $1, ai_error_msg = NULL, updated_at = NOW()
			WHERE id = $2`,
			masterCard.ID, id)

		if err != nil {
			log.Printf("❌ Ошибка обновления статуса: %v\n", err)
		} else {
			fmt.Printf("✅ Объявление %d успешно привязано к официальной карте: %s\n", id, masterCard.Name)
		}

		// 5. Жесткий лимит Google (5 запросов в минуту = 1 запрос каждые 12 секунд)
		time.Sleep(12 * time.Second)
	}
}
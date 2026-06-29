package workers

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
	"pokechart-api/internal/services/vision"
)

func StartVisionWorker(db *sql.DB, cardRepo *repository.CardRepo) {
	fmt.Println("🤖 Фоновый AI-воркер запущен и ждет задачи...")

	for {
		var id int
		var imageUrl string

		// SKIP LOCKED — берём одну задачу без блокировки параллельных воркеров.
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
			time.Sleep(10 * time.Second)
			continue
		} else if err != nil {
			log.Printf("⚠️ Ошибка БД при чтении очереди: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Printf("🔍 Воркер взял в работу объявление ID: %d...\n", id)

		aiResult, err := vision.RecognizeCard(imageUrl)

		if err != nil {
			// Возвращаем в pending, чтобы одна «проблемная» карточка не блокировала очередь.
			log.Printf("❌ Ошибка ИИ: %v. Откладываем задачу...", err)
			db.Exec("UPDATE market_listings SET status = 'pending', ai_error_msg = $1, updated_at = NOW() WHERE id = $2", err.Error(), id)
			time.Sleep(5 * time.Second)
			continue
		}

		// для дебага и будущего матчинга
		if aiResult.CardNo != "" {
			db.Exec("UPDATE market_listings SET ai_card_no = $1, updated_at = NOW() WHERE id = $2", aiResult.CardNo, id)
		}
		if aiResult.Series != "" {
			db.Exec("UPDATE market_listings SET ai_series = $1, updated_at = NOW() WHERE id = $2", aiResult.Series, id)
		}

		if !aiResult.IsPokemonCard {
			db.Exec("UPDATE market_listings SET status = 'rejected', ai_error_msg = 'Не покемон', updated_at = NOW() WHERE id = $1", id)
			continue
		}

		// Стратегия поиска: card_no → name+series (с визуальным выбором при коллизиях) → name+rarity.
		var masterCard *models.Card
		if aiResult.CardNo != "" {
			if c, e := cardRepo.FindBestByCardNoThenName(aiResult.CardNo, aiResult.Name); e == nil {
				masterCard = c
			}
		}
		if masterCard == nil && aiResult.Series != "" {
			cands, _ := cardRepo.FindCandidateCardsByNameAndSeries(aiResult.Name, aiResult.Series, 5)
			if len(cands) == 1 {
				cc := cands[0]
				masterCard = &cc
			} else if len(cands) > 1 {
				if cc, e := vision.ResolveMasterCardByImages(imageUrl, cands); e == nil && cc != nil {
					masterCard = cc
				}
			}
			// If we still didn't resolve, try a simple DB pick as last attempt within this stage.
			if masterCard == nil {
				if c, e := cardRepo.FindCardByNameAndSeries(aiResult.Name, aiResult.Series); e == nil {
					masterCard = c
				}
			}
		}
		// fallback: имя + редкость
		if masterCard == nil {
			masterCard, err = cardRepo.FindCardByNameAndRarity(aiResult.Name, aiResult.Rarity)
		} else {
			err = nil
		}
		if err != nil {
			log.Printf("⚠️ Карта %s %s не найдена в БД\n", aiResult.Name, aiResult.Rarity)
			db.Exec("UPDATE market_listings SET status = 'error', ai_error_msg = 'Нет в справочнике', updated_at = NOW() WHERE id = $1", id)
			continue
		}

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

		// Gemini Free: лимит 5 RPM → минимум 12 секунд между запросами.
		time.Sleep(12 * time.Second)
	}
}
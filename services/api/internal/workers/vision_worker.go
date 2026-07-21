package workers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
	"pokechart-api/internal/services/vision"
)

type VisionWorker struct {
	db       *sql.DB
	cardRepo *repository.CardRepo
}

func NewVisionWorker(db *sql.DB, cardRepo *repository.CardRepo) *VisionWorker {
	return &VisionWorker{db: db, cardRepo: cardRepo}
}

func (w *VisionWorker) Start(ctx context.Context) {
	fmt.Println("🤖 Vision-воркер запущен...")
	for {
		id, imageURL, price, source, ok := w.claimNext()
		if !ok {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 Vision-воркер остановлен")
				return
			case <-time.After(10 * time.Second):
				continue
			}
		}

		w.process(id, imageURL, price, source)

		// Gemini Free: 5 RPM → минимум 12 секунд между запросами
		select {
		case <-ctx.Done():
			fmt.Println("🛑 Vision-воркер остановлен")
			return
		case <-time.After(12 * time.Second):
		}
	}
}

func (w *VisionWorker) claimNext() (id int, imageURL string, price int, source string, ok bool) {
	const q = `
		UPDATE market_listings
		SET status = 'processing', updated_at = NOW()
		WHERE id = (
			SELECT id FROM market_listings
			WHERE status = 'pending'
			ORDER BY updated_at ASC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, image_url, COALESCE(price, 0), COALESCE(source, 'daangn')
	`
	err := w.db.QueryRow(q).Scan(&id, &imageURL, &price, &source)
	if err == sql.ErrNoRows {
		return 0, "", 0, "", false
	}
	if err != nil {
		log.Printf("⚠️ Ошибка чтения очереди: %v", err)
		return 0, "", 0, "", false
	}
	return id, imageURL, price, source, true
}

func (w *VisionWorker) process(id int, imageURL string, price int, source string) {
	fmt.Printf("🔍 Обрабатываю ID=%d source=%s\n", id, source)

	aiResult, err := vision.RecognizeCard(imageURL)
	if err != nil {
		log.Printf("❌ Ошибка AI (ID=%d): %v. Откладываю...", id, err)
		w.exec(
			"UPDATE market_listings SET status='pending', ai_error_msg=$1, updated_at=NOW() WHERE id=$2",
			err.Error(), id,
		)
		return
	}

	w.exec(`
		UPDATE market_listings
		SET ai_card_no = CASE WHEN $1 != '' THEN $1 ELSE ai_card_no END,
		    ai_series  = CASE WHEN $2 != '' THEN $2 ELSE ai_series  END,
		    updated_at = NOW()
		WHERE id = $3
	`, aiResult.CardNo, aiResult.Series, id)

	if !aiResult.IsPokemonCard {
		w.exec(
			"UPDATE market_listings SET status='rejected', ai_error_msg='Не покемон', updated_at=NOW() WHERE id=$1",
			id,
		)
		return
	}

	masterCard := w.findMasterCard(imageURL, aiResult)
	if masterCard == nil {
		log.Printf("⚠️ Карта '%s' (%s) не найдена в справочнике (ID=%d)", aiResult.Name, aiResult.Rarity, id)
		w.exec(
			"UPDATE market_listings SET status='error', ai_error_msg='Нет в справочнике', updated_at=NOW() WHERE id=$1",
			id,
		)
		return
	}

	w.exec(`
		UPDATE market_listings
		SET status='active', master_card_id=$1, ai_error_msg=NULL, updated_at=NOW()
		WHERE id=$2
	`, masterCard.ID, id)

	w.recordPriceSnapshot(masterCard.ID, id, price, source)

	fmt.Printf("✅ ID=%d привязан к карте: %s\n", id, masterCard.Name)
}

// findMasterCard пробует три стратегии: card_no → name+series → name+rarity.
func (w *VisionWorker) findMasterCard(imageURL string, ai *vision.VisionResult) *models.Card {
	if ai.CardNo != "" {
		if c, err := w.cardRepo.FindBestByCardNoThenName(ai.CardNo, ai.Name); err == nil {
			return c
		}
	}

	if ai.Series != "" {
		cands, _ := w.cardRepo.FindCandidateCardsByNameAndSeries(ai.Name, ai.Series, 5)
		switch {
		case len(cands) == 1:
			cc := cands[0]
			return &cc
		case len(cands) > 1:
			if cc, err := vision.ResolveMasterCardByImages(imageURL, cands); err == nil && cc != nil {
				return cc
			}
			if c, err := w.cardRepo.FindCardByNameAndSeries(ai.Name, ai.Series); err == nil {
				return c
			}
		}
	}

	if c, err := w.cardRepo.FindCardByNameAndRarity(ai.Name, ai.Rarity); err == nil {
		return c
	}

	return nil
}

func (w *VisionWorker) recordPriceSnapshot(cardID string, listingID, price int, source string) {
	w.exec(`
		INSERT INTO price_snapshots (card_id, listing_id, price, source)
		VALUES ($1, $2, $3, $4)
	`, cardID, listingID, price, source)
}

func (w *VisionWorker) exec(query string, args ...any) {
	if _, err := w.db.Exec(query, args...); err != nil {
		log.Printf("⚠️ db.Exec: %v | query: %.120s", err, query)
	}
}

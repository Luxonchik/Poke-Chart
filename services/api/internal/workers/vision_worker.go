package workers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"pokechart-api/internal/models"
	"pokechart-api/internal/repository"
	"pokechart-api/internal/services/vision"
)

const (
	// Gemini Free: 5 RPM → at least 12 seconds between requests.
	requestInterval = 12 * time.Second
	busyBackoffMin  = 30 * time.Second
	busyBackoffMax  = 5 * time.Minute
	// A listing that fails this many times for reasons other than Gemini load gets status 'error'.
	maxAIAttempts = 3
)

// attemptPrefix matches the "[2/3] " counter kept at the start of ai_error_msg.
var attemptPrefix = regexp.MustCompile(`^\[(\d+)/\d+\] `)

type VisionWorker struct {
	db       *sql.DB
	cardRepo *repository.CardRepo
	backoff  time.Duration // current pause after Gemini 429/503; 0 when healthy
}

type queuedListing struct {
	ID        int
	ImageURL  string
	Price     int
	Source    string
	LastError string
}

func NewVisionWorker(db *sql.DB, cardRepo *repository.CardRepo) *VisionWorker {
	return &VisionWorker{db: db, cardRepo: cardRepo}
}

func (w *VisionWorker) Start(ctx context.Context) {
	fmt.Println("🤖 Vision-воркер запущен...")
	for {
		listing, ok := w.claimNext()
		if !ok {
			select {
			case <-ctx.Done():
				fmt.Println("🛑 Vision-воркер остановлен")
				return
			case <-time.After(10 * time.Second):
				continue
			}
		}

		delay := requestInterval
		if w.process(listing) {
			w.backoff = min(max(w.backoff*2, busyBackoffMin), busyBackoffMax)
			delay = w.backoff
			fmt.Printf("⏸ Gemini перегружен, пауза %s\n", delay)
		} else {
			w.backoff = 0
		}

		select {
		case <-ctx.Done():
			fmt.Println("🛑 Vision-воркер остановлен")
			return
		case <-time.After(delay):
		}
	}
}

func (w *VisionWorker) claimNext() (queuedListing, bool) {
	const q = `
		UPDATE market_listings
		SET status = 'processing', updated_at = NOW()
		WHERE id = (
			SELECT id FROM market_listings
			WHERE status = 'pending'
			ORDER BY updated_at ASC, created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, image_url, COALESCE(price, 0), COALESCE(source, 'daangn'), COALESCE(ai_error_msg, '')
	`
	var l queuedListing
	err := w.db.QueryRow(q).Scan(&l.ID, &l.ImageURL, &l.Price, &l.Source, &l.LastError)
	if err == sql.ErrNoRows {
		return l, false
	}
	if err != nil {
		log.Printf("⚠️ Ошибка чтения очереди: %v", err)
		return l, false
	}
	return l, true
}

// process recognises one listing and reports whether Gemini was overloaded.
func (w *VisionWorker) process(l queuedListing) (busy bool) {
	id, imageURL, price, source := l.ID, l.ImageURL, l.Price, l.Source

	aiResult, err := vision.RecognizeCard(imageURL)
	if errors.Is(err, vision.ErrModelBusy) {
		fmt.Printf("🔍 ID=%d → отложено: %v\n", id, err)
		w.exec(
			"UPDATE market_listings SET status='pending', updated_at=NOW() WHERE id=$1",
			id,
		)
		return true
	}
	if err != nil {
		w.failAttempt(l, err)
		return false
	}

	w.exec(`
		UPDATE market_listings
		SET ai_card_no = CASE WHEN $1 != '' THEN $1 ELSE ai_card_no END,
		    ai_series  = CASE WHEN $2 != '' THEN $2 ELSE ai_series  END,
		    updated_at = NOW()
		WHERE id = $3
	`, aiResult.CardNo, aiResult.Series, id)

	if !aiResult.IsPokemonCard {
		fmt.Printf("🔍 ID=%d → rejected: не одиночная карта\n", id)
		w.exec(
			"UPDATE market_listings SET status='rejected', ai_error_msg='Не покемон', updated_at=NOW() WHERE id=$1",
			id,
		)
		return false
	}

	masterCard := w.findMasterCard(imageURL, aiResult)
	if masterCard == nil {
		fmt.Printf("🔍 ID=%d → error: нет в справочнике  (AI: %s)\n", id, describe(aiResult))
		w.exec(
			"UPDATE market_listings SET status='error', ai_error_msg='Нет в справочнике', updated_at=NOW() WHERE id=$1",
			id,
		)
		return false
	}

	w.exec(`
		UPDATE market_listings
		SET status='active', master_card_id=$1, ai_error_msg=NULL, updated_at=NOW()
		WHERE id=$2
	`, masterCard.ID, id)

	w.recordPriceSnapshot(masterCard.ID, id, price, source)

	fmt.Printf("🔍 ID=%d → active: %s %s  (AI: %s)\n", id, masterCard.Name, masterCard.CardNo, describe(aiResult))
	return false
}

// failAttempt counts a non-load failure in ai_error_msg ("[n/max] reason") and gives up after maxAIAttempts.
func (w *VisionWorker) failAttempt(l queuedListing, cause error) {
	attempt := 1
	if m := attemptPrefix.FindStringSubmatch(l.LastError); m != nil {
		prev, _ := strconv.Atoi(m[1])
		attempt = prev + 1
	}

	status := "pending"
	if attempt >= maxAIAttempts {
		status = "error"
	}
	msg := fmt.Sprintf("[%d/%d] %v", attempt, maxAIAttempts, cause)
	fmt.Printf("🔍 ID=%d → %s, попытка %d/%d: %v\n", l.ID, status, attempt, maxAIAttempts, cause)

	w.exec(
		"UPDATE market_listings SET status=$1, ai_error_msg=$2, updated_at=NOW() WHERE id=$3",
		status, msg, l.ID,
	)
}

func describe(r *vision.VisionResult) string {
	return fmt.Sprintf("%s · %s · %s", r.Name, r.Rarity, r.CardNo)
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

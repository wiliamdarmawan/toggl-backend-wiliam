package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"toggl-test-wiliam/model"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CreateDeckSerializer struct {
	ID        string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
}

type OpenDeckSerializer struct {
	ID        string       `json:"deck_id"`
	Shuffled  bool         `json:"shuffled"`
	Remaining int          `json:"remaining"`
	Cards     []model.Card `json:"cards"`
}

func connectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("game.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func CreateNewDeck(w http.ResponseWriter, r *http.Request) {
	db, ok := r.Context().Value("db").(*gorm.DB)
	if !ok {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	shuffleParam := r.URL.Query().Get("shuffle")
	cardsParam := r.URL.Query().Get("cards")

	// Default to not shuffling
	shuffle := false
	var cards []string

	// Parse "cards" query parameter
	if cardsParam != "" {
		cards = strings.Split(cardsParam, ",")
		validCards := getValidCards("FRENCH", cards, db)
		invalidCards := getInvalidCards(cards, validCards)

		if len(invalidCards) > 0 {
			http.Error(w, fmt.Sprintf("invalid cards: %v", invalidCards), http.StatusBadRequest)
			return
		}
	} else {
		db.Model(&model.Card{}).Where("card_type = ?", "FRENCH").Pluck("code", &cards)
	}

	// Parse "shuffle" query parameter
	if shuffleParam != "" {
		shuffle, _ = strconv.ParseBool(shuffleParam)
	}

	var err error
	deck := model.Deck{}
	deck, err = deck.Create(cards)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if shuffle {
		deck.Shuffle()
	}

	db.Create(&deck)
	w.Header().Set("Content-Type", "application/json")
	response := CreateDeckSerializer{
		ID:        deck.ID,
		Shuffled:  deck.Shuffled,
		Remaining: len(deck.Cards),
	}

	json.NewEncoder(w).Encode(response)
}

func OpenDeck(w http.ResponseWriter, r *http.Request) {
	db, ok := r.Context().Value("db").(*gorm.DB)
	if !ok {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	deck := model.Deck{}
	deck_id := mux.Vars(r)["deck_id"]

	db.First(&deck, "id = ?", deck_id)
	if deck.ID == "" {
		http.Error(w, "Deck not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := OpenDeckSerializer{
		ID:        deck.ID,
		Shuffled:  deck.Shuffled,
		Remaining: len(deck.Cards),
		Cards:     deck.Cards,
	}

	json.NewEncoder(w).Encode(response)
}

func DrawCards(w http.ResponseWriter, r *http.Request) {
	db, ok := r.Context().Value("db").(*gorm.DB)
	if !ok {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	deck := model.Deck{}
	deck_id := mux.Vars(r)["deck_id"]

	db.First(&deck, "id = ?", deck_id)
	if deck.ID == "" {
		http.Error(w, "Deck not found", http.StatusNotFound)
		return
	}

	countParam := r.URL.Query().Get("count")
	count, _ := strconv.Atoi(countParam)
	if count == 0 {
		count = 1
	}

	if count > len(deck.Cards) {
		http.Error(w, "Not enough cards in the deck", http.StatusBadRequest)
		return
	}

	cards, err := deck.Draw(count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db.Save(&deck)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

func getValidCards(card_type string, codes []string, db *gorm.DB) map[string]string {
	var validCards []string
	db.Model(&model.Card{}).Where("card_type = ? AND code IN (?)", card_type, codes).Pluck("code", &validCards)

	validCardsMap := make(map[string]string)
	for _, card := range validCards {
		validCardsMap[card] = card
	}

	return validCardsMap
}

func getInvalidCards(codes []string, validCards map[string]string) []string {
	var invalidCards []string
	for _, card := range codes {
		if validCards[card] == "" {
			invalidCards = append(invalidCards, card)
		}
	}
	return invalidCards
}

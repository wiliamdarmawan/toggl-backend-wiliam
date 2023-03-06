package main_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	api "toggl-test-wiliam/api"
	model "toggl-test-wiliam/model"
	seeds "toggl-test-wiliam/seeds"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIntegration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Seeds the database with a full deck of cards
	if err = db.AutoMigrate(&model.Card{}, &model.Deck{}); err == nil && db.Migrator().HasTable(&model.Card{}) {
		if err := db.First(&model.Card{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			seeds.FrenchCardDeck(db)
		}
	}

	dbMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pass the database connection to the context
			ctx := context.WithValue(r.Context(), "db", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r := mux.NewRouter()
	r.Use(dbMiddleware)

	r.HandleFunc("/deck", api.CreateNewDeck).Methods("POST")
	r.HandleFunc("/deck/{deck_id}", api.OpenDeck).Methods("GET")
	r.HandleFunc("/deck/{deck_id}/draw", api.DrawCards).Methods("GET")

	ts := httptest.NewServer(r)
	defer ts.Close()

	var count int64
	db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(0), "Deck shouldn't be created yet at this point")

	// Test POST /deck endpoint
	resp, err := http.Post(ts.URL+"/deck", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(1), "Deck should be created at this point")

	deck := api.CreateDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.NotEmpty(t, deck.ID)
	assert.Equal(t, 52, deck.Remaining)

	// Test GET /deck/{deck_id} endpoint
	resp, err = http.Get(ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck2 := api.OpenDeckSerializer{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck2))
	assert.Equal(t, deck.ID, deck2.ID)
	assert.Equal(t, deck.Remaining, deck2.Remaining)

	// Test GET /deck/{deck_id}/draw endpoint
	resp, err = http.Get(ts.URL + "/deck/" + deck.ID + "/draw")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	beforeDrawn := deck.Remaining
	cards := []model.Card{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&cards))
	assert.Len(t, cards, 1)

	resp, _ = http.Get(ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.Equal(t, len(cards)+deck.Remaining, beforeDrawn)

	db.Migrator().DropTable(&model.Card{}, &model.Deck{})
}

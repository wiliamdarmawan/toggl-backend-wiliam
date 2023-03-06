package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"toggl-test-wiliam/api"
	"toggl-test-wiliam/model"
	"toggl-test-wiliam/seeds"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("game.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

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

	fmt.Println("Listening on port 80....")
	http.ListenAndServe(":80", r)
}

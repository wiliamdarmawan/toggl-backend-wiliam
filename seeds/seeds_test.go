package seeds_test

import (
	"errors"
	"testing"

	model "toggl-test-wiliam/model"
	"toggl-test-wiliam/seeds"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFrenchCardDeck(t *testing.T) {
	// Set up a test database connection
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database connection: %v", err)
	}

	// Migrate and seed data
	if err = db.AutoMigrate(&model.Card{}, &model.Deck{}); err == nil && db.Migrator().HasTable(&model.Card{}) {
		if err := db.First(&model.Card{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			seeds.FrenchCardDeck(db)
		}
	}

	// Retrieve all cards from the database
	var cards []model.Card
	db.Model(&model.Card{}).Find(&cards)

	suits := []string{"CLUB", "DIAMONDS", "HEARTS", "SPADES"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

	// Assert that the number of cards retrieved is equal to the expected number
	expectedNumCards := len(suits) * len(values)
	assert.Equal(t, expectedNumCards, len(cards))

	codes := []string{}
	// Assert that each card created by the FrenchCardDeck function is present in the database
	for _, suit := range suits {
		for _, value := range values {
			code := value + suit[0:1]
			codes = append(codes, code)
		}
	}

	for _, card := range cards {
		assert.Contains(t, codes, card.Code)
	}
}

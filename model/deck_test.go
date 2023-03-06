package model_test

import (
	"testing"
	"toggl-test-wiliam/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDeck_TooFew(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{}

	_, err := deck.Create(cardCodes)
	assert.EqualError(t, err, "too few cards provided")
}

func TestCreateDeck_TooMany(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{}

	for i := 0; i < 53; i++ {
		cardCodes = append(cardCodes, "AS")
	}

	_, err := deck.Create(cardCodes)
	assert.EqualError(t, err, "too many cards provided")
}

func TestCreateDeck_Valid(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{"AS", "2S", "3S", "4S", "5S", "6S", "7S", "8S", "9S", "10S", "JS", "QS", "KS"}

	createdDeck, err := deck.Create(cardCodes)
	assert.NoError(t, err)
	assert.Equal(t, createdDeck.Remaining, len(cardCodes))

	var createdDeckCodes []string
	for _, card := range createdDeck.Cards {
		createdDeckCodes = append(createdDeckCodes, card.Code)
	}
	assert.Equal(t, createdDeckCodes, cardCodes)
}

func TestDrawCards_TooMany(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{"AS", "2S", "3S", "4S", "5S", "6S", "7S", "8S", "9S", "10S", "JS", "QS", "KS"}

	createdDeck, err := deck.Create(cardCodes)
	assert.NoError(t, err)

	_, err = createdDeck.Draw(14)
	assert.EqualError(t, err, "too many cards requested")
}

func TestDrawCards_Valid(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{"AS", "2S", "3S", "4S", "5S", "6S", "7S", "8S", "9S", "10S", "JS", "QS", "KS"}

	createdDeck, err := deck.Create(cardCodes)
	assert.NoError(t, err)

	drawnCards, err := createdDeck.Draw(5)
	assert.NoError(t, err)
	assert.Equal(t, len(drawnCards), 5)
	assert.Equal(t, createdDeck.Remaining, len(cardCodes)-len(drawnCards))

	var drawnCardsCodes []string
	for _, card := range drawnCards {
		drawnCardsCodes = append(drawnCardsCodes, card.Code)
	}
	assert.Equal(t, drawnCardsCodes, cardCodes[len(cardCodes)-len(drawnCardsCodes):])
}

func TestShuffle(t *testing.T) {
	deck := model.Deck{}
	cardCodes := []string{"AS", "KS", "QS", "JS", "10S", "9S", "8S", "7S", "6S", "5S", "4S", "3S", "2S"}
	createdDeck, _ := deck.Create(cardCodes)
	createdDeck.Shuffle()
	assert.True(t, createdDeck.Shuffled)

	var createdDeckCodes []string
	for _, card := range createdDeck.Cards {
		createdDeckCodes = append(createdDeckCodes, card.Code)
	}
	assert.NotEqual(t, createdDeckCodes, cardCodes)
}

func TestDeck_BeforeSave(t *testing.T) {
	deck := model.Deck{}
	deck.Cards = []model.Card{
		{Code: "AC", Suit: "CLUBS", Value: "ACE", CardType: "FRENCH"},
		{Code: "KD", Suit: "DIAMONDS", Value: "KING", CardType: "FRENCH"},
	}
	err := deck.BeforeSave(nil)
	require.NoError(t, err)

	expectedJSON := `[
  {
    "value": "ACE",
    "suit": "CLUBS",
    "code": "AC"
  },
  {
    "value": "KING",
    "suit": "DIAMONDS",
    "code": "KD"
  }
]`
	assert.JSONEq(t, expectedJSON, string(deck.CardsJSON))
}

func TestDeck_AfterFind(t *testing.T) {
	expectedJSON := `[
  {
    "value": "ACE",
    "suit": "CLUBS",
    "code": "AC"
  },
  {
    "value": "KING",
    "suit": "DIAMONDS",
    "code": "KD"
  }
]`

	deck := model.Deck{CardsJSON: []byte(expectedJSON)}
	err := deck.AfterFind(nil)
	require.NoError(t, err)

	expectedCards := []model.Card{
		{Code: "AC", Suit: "CLUBS", Value: "ACE"},
		{Code: "KD", Suit: "DIAMONDS", Value: "KING"},
	}
	assert.Equal(t, expectedCards, deck.Cards)
}

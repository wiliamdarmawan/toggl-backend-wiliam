package seeds

import (
	model "toggl-test-wiliam/model"

	"gorm.io/gorm"
)

var suitsList = map[string][]string{
	"FRENCH": {"CLUB", "DIAMONDS", "HEARTS", "SPADES"},
}

var valuesList = map[string][]string{
	"FRENCH": {"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"},
}

func FrenchCardDeck(db *gorm.DB) {
	cards := []model.Card{}
	suits := suitsList["FRENCH"]
	values := valuesList["FRENCH"]

	for _, suit := range suits {
		for _, value := range values {
			card := model.Card{
				Value:    value,
				Suit:     suit,
				Code:     value + suit[0:1],
				CardType: "FRENCH",
			}
			cards = append(cards, card)
		}
	}

	db.Create(&cards)
}

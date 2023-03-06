package model

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Deck struct {
	gorm.Model
	ID        string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
	CardsJSON []byte `json:"cards_json" gorm:"column:cards"`
	Cards     []Card `json:"cards" gorm:"-"`
}

type Card struct {
	Value    string `json:"value"`
	Suit     string `json:"suit"`
	Code     string `json:"code"`
	CardType string `json:"-"`
}

var maxCards = map[string]int{
	"FRENCH": 52,
}
var minCards = map[string]int{
	"FRENCH": 1,
}

var valueNames = map[string]string{
	"A":  "ACE",
	"2":  "2",
	"3":  "3",
	"4":  "4",
	"5":  "5",
	"6":  "6",
	"7":  "7",
	"8":  "8",
	"9":  "9",
	"10": "10",
	"J":  "JACK",
	"Q":  "QUEEN",
	"K":  "KING",
}

var suitNames = map[string]string{
	"C": "CLUBS",
	"D": "DIAMONDS",
	"H": "HEARTS",
	"S": "SPADES",
}

func (d *Deck) Create(cardCodes []string) (Deck, error) {
	if len(cardCodes) < minCards["FRENCH"] {
		return Deck{}, errors.New("too few cards provided")
	} else if len(cardCodes) > maxCards["FRENCH"] {
		return Deck{}, errors.New("too many cards provided")
	}

	deck := Deck{}
	deck.ID = uuid.New().String()

	for _, code := range cardCodes {
		value := ""
		suit := ""

		if len(code) <= 2 {
			value = code[0:1]
			suit = code[1:2]
		} else {
			value = code[0:2]
			suit = code[2:3]
		}

		card := Card{
			Code:     code,
			Suit:     suitNames[suit],
			Value:    valueNames[value],
			CardType: "FRENCH",
		}
		deck.Cards = append(deck.Cards, card)
	}
	deck.Remaining = len(deck.Cards)

	return deck, nil
}

func (d *Deck) Draw(count int) ([]Card, error) {
	if count > d.Remaining {
		return []Card{}, errors.New("too many cards requested")
	}

	drawnCards := d.Cards[len(d.Cards)-count:]
	d.Cards = d.Cards[0 : len(d.Cards)-count]
	d.Remaining = len(d.Cards)

	return drawnCards, nil
}

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})

	d.Shuffled = true
}

// Implement BeforeSave hook to encode Cards field to JSON
func (d *Deck) BeforeSave(*gorm.DB) error {
	var err error
	if len(d.Cards) > 0 {
		d.CardsJSON, err = json.Marshal(d.Cards)
		if err != nil {
			return err
		}
	}
	return nil
}

// Implement AfterFind hook to decode Cards field from JSON
func (d *Deck) AfterFind(*gorm.DB) error {
	if len(d.CardsJSON) > 0 {
		err := json.Unmarshal(d.CardsJSON, &d.Cards)
		if err != nil {
			return err
		}
	}
	return nil
}

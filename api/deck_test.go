package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	api "toggl-test-wiliam/api"
	model "toggl-test-wiliam/model"
	seeds "toggl-test-wiliam/seeds"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type APITestSuite struct {
	suite.Suite
	db *gorm.DB
	ts *httptest.Server
}

func (suite *APITestSuite) SetupTest() {
	var err error
	suite.db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Seeds the database with a full deck of cards
	if err = suite.db.AutoMigrate(&model.Card{}, &model.Deck{}); err == nil && suite.db.Migrator().HasTable(&model.Card{}) {
		if err := suite.db.First(&model.Card{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			seeds.FrenchCardDeck(suite.db)
		}
	}

	dbMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pass the database connection to the context
			ctx := context.WithValue(r.Context(), "db", suite.db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// create a test router and add handlers to it
	r := mux.NewRouter()
	r.Use(dbMiddleware)

	r.HandleFunc("/deck", api.CreateNewDeck).Methods("POST")
	r.HandleFunc("/deck/{deck_id}", api.OpenDeck).Methods("GET")
	r.HandleFunc("/deck/{deck_id}/draw", api.DrawCards).Methods("GET")

	suite.ts = httptest.NewServer(r)
}

func (suite *APITestSuite) TearDownTest() {
	// cleanup
	suite.ts.Close()
	suite.db.Exec("DROP TABLE IF EXISTS decks;")
	suite.db.Exec("DROP TABLE IF EXISTS cards;")
}

func TestCreateNewDeck_WithNoParameters(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	var count int64
	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(0), "Deck shouldn't be created yet at this point")

	// create a test HTTP request without parameters
	resp, err := http.Post(testSuite.ts.URL+"/deck", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck := api.CreateDeckSerializer{}

	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(1), "Deck should be created at this point")

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.NotEmpty(t, deck.ID)
	assert.Equal(t, deck.Shuffled, false)
	assert.Equal(t, deck.Remaining, 52)

	// Open Deck to test if the cards are correctly created
	resp, err = http.Get(testSuite.ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	openedDeck := api.OpenDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&openedDeck))
	assert.Equal(t, openedDeck.ID, deck.ID)
	assert.Equal(t, openedDeck.Shuffled, false)

	var cards []string
	testSuite.db.Model(model.Card{}).Pluck("code", &cards)

	var codes []string
	for _, card := range openedDeck.Cards {
		codes = append(codes, card.Code)
	}
	assert.Equal(t, cards, codes)

	testSuite.TearDownTest()
}

func TestCreateNewDeck_WithShuffledParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	var count int64
	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(0), "Deck shouldn't be created yet at this point")

	// create a test HTTP request with shuffle parameter as true
	resp, err := http.Post(testSuite.ts.URL+"/deck?shuffle=true", "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck := api.CreateDeckSerializer{}

	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(1), "Deck should be created at this point")

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.NotEmpty(t, deck.ID)
	assert.Equal(t, deck.Shuffled, true)
	assert.Equal(t, deck.Remaining, 52)

	// Open Deck to test if the cards are correctly created
	resp, err = http.Get(testSuite.ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	openedDeck := api.OpenDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&openedDeck))
	assert.Equal(t, openedDeck.ID, deck.ID)
	assert.Equal(t, openedDeck.Shuffled, true)
	assert.Equal(t, openedDeck.Remaining, 52)

	var cards []string
	testSuite.db.Model(model.Card{}).Pluck("code", &cards)

	var codes []string
	for _, card := range openedDeck.Cards {
		codes = append(codes, card.Code)
	}
	assert.NotEqual(t, cards, codes)

	testSuite.TearDownTest()
}

func TestCreateNewDeck_WithInvalidCardsParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	resp, _ := http.Post(testSuite.ts.URL+"/deck?cards=XXX,YYY,AC", "application/json", nil)
	resp_body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "invalid cards: [XXX YYY]\n", string(resp_body))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	testSuite.TearDownTest()
}

func TestCreateNewDeck_WithCardsParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	var count int64
	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(0), "Deck shouldn't be created yet at this point")

	// create a test HTTP request with cards parameters
	cardParams := "AS,2S,3S,4S,5S,6S,7S,8S,9S,10S,JS,QS,KS"
	resp, err := http.Post(testSuite.ts.URL+"/deck?cards="+cardParams, "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck := api.CreateDeckSerializer{}
	cards := strings.Split(cardParams, ",")

	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(1), "Deck should be created at this point")

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.NotEmpty(t, deck.ID)
	assert.Equal(t, deck.Shuffled, false)
	assert.Equal(t, deck.Remaining, len(cards))

	// Open Deck to test if the cards are correctly created
	resp, err = http.Get(testSuite.ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	openedDeck := api.OpenDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&openedDeck))
	assert.Equal(t, openedDeck.ID, deck.ID)
	assert.Equal(t, openedDeck.Shuffled, false)
	assert.Equal(t, openedDeck.Remaining, len(cards))

	var codes []string
	for _, card := range openedDeck.Cards {
		codes = append(codes, card.Code)
	}
	assert.Equal(t, cards, codes)
	testSuite.TearDownTest()
}

func TestCreateNewDeck_WithShuffledAndCardsParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	suite.Run(t, testSuite)
	testSuite.SetupTest()

	var count int64
	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(0), "Deck shouldn't be created yet at this point")

	// create a test HTTP request with shuffle and cards query parameters
	cardsParam := "AH,2H,3H,4H,5H"
	resp, err := http.Post(testSuite.ts.URL+"/deck?shuffle=true&cards="+cardsParam, "application/json", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck := api.CreateDeckSerializer{}
	cards := strings.Split(cardsParam, ",")

	testSuite.db.Model(model.Deck{}).Count(&count)
	assert.Equal(t, count, int64(1), "Deck should be created at this point")

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.NotEmpty(t, deck.ID)
	assert.Equal(t, deck.Shuffled, true)
	assert.Equal(t, deck.Remaining, len(cards))

	// Open Deck to test if the cards are shuffled
	resp, err = http.Get(testSuite.ts.URL + "/deck/" + deck.ID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	openedDeck := api.OpenDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&openedDeck))
	assert.Equal(t, openedDeck.ID, deck.ID)
	assert.Equal(t, openedDeck.Shuffled, true)
	assert.Equal(t, openedDeck.Remaining, len(cards))

	var codes []string
	for _, card := range openedDeck.Cards {
		codes = append(codes, card.Code)
	}
	assert.NotEqual(t, cards, codes)
	testSuite.TearDownTest()
}

func TestOpenDeck_Failed(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	resp, _ := http.Get(testSuite.ts.URL + "/deck/zxczxczxc")
	resp_body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Deck not found\n", string(resp_body))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	testSuite.TearDownTest()
}

func TestOpenDeck_Success(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	mockDeck := &model.Deck{
		ID:        "test_deck_id",
		Shuffled:  true,
		Remaining: 4,
		Cards: []model.Card{
			{Value: "ACE", Suit: "CLUBS", Code: "AC"},
			{Value: "ACE", Suit: "DIAMONDS", Code: "AD"},
			{Value: "ACE", Suit: "HEARTS", Code: "AH"},
			{Value: "ACE", Suit: "SPADES", Code: "AS"},
		},
	}
	testSuite.db.Create(mockDeck)
	resp, err := http.Get(testSuite.ts.URL + "/deck/test_deck_id")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	deck := api.OpenDeckSerializer{}

	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&deck))
	assert.Equal(t, deck.ID, mockDeck.ID)
	assert.Equal(t, deck.Shuffled, mockDeck.Shuffled)
	assert.Equal(t, deck.Remaining, len(mockDeck.Cards))
	assert.Equal(t, deck.Cards, mockDeck.Cards)

	testSuite.TearDownTest()
}

func TestDrawCards_NotFound(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	resp, _ := http.Get(testSuite.ts.URL + "/deck/zxczxczxc/draw")
	resp_body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Deck not found\n", string(resp_body))
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	testSuite.TearDownTest()
}

func TestDrawCards_NotEnoughCards(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	mockDeck := &model.Deck{
		ID:        "test_deck_id",
		Shuffled:  true,
		Remaining: 4,
		Cards: []model.Card{
			{Value: "ACE", Suit: "CLUBS", Code: "AC"},
			{Value: "ACE", Suit: "DIAMONDS", Code: "AD"},
			{Value: "ACE", Suit: "HEARTS", Code: "AH"},
			{Value: "ACE", Suit: "SPADES", Code: "AS"},
		},
	}
	testSuite.db.Create(mockDeck)
	resp, _ := http.Get(testSuite.ts.URL + "/deck/test_deck_id/draw?count=5")
	resp_body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "Not enough cards in the deck\n", string(resp_body))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	testSuite.TearDownTest()
}

func TestDrawCards_WithoutParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	mockDeck := &model.Deck{
		ID:        "test_deck_id",
		Shuffled:  true,
		Remaining: 4,
		Cards: []model.Card{
			{Value: "ACE", Suit: "CLUBS", Code: "AC"},
			{Value: "ACE", Suit: "DIAMONDS", Code: "AD"},
			{Value: "ACE", Suit: "HEARTS", Code: "AH"},
			{Value: "ACE", Suit: "SPADES", Code: "AS"},
		},
	}
	testSuite.db.Create(mockDeck)

	var deck model.Deck
	testSuite.db.First(&deck, "id = ?", "test_deck_id")
	cardsCount := deck.Remaining
	assert.Equal(t, cardsCount, 4)

	resp, err := http.Get(testSuite.ts.URL + "/deck/test_deck_id/draw")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	testSuite.db.First(&deck, "id = ?", "test_deck_id")
	assert.Equal(t, cardsCount-1, deck.Remaining)

	drawnCard := []model.Card{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&drawnCard))
	assert.Equal(t, drawnCard, mockDeck.Cards[len(mockDeck.Cards)-1:])
}

func TestDrawCards_WithCountParameter(t *testing.T) {
	testSuite := new(APITestSuite)
	testSuite.SetupTest()

	mockDeck := &model.Deck{
		ID:        "test_deck_id",
		Shuffled:  true,
		Remaining: 4,
		Cards: []model.Card{
			{Value: "ACE", Suit: "CLUBS", Code: "AC"},
			{Value: "ACE", Suit: "DIAMONDS", Code: "AD"},
			{Value: "ACE", Suit: "HEARTS", Code: "AH"},
			{Value: "ACE", Suit: "SPADES", Code: "AS"},
		},
	}
	testSuite.db.Create(mockDeck)

	var deck model.Deck
	testSuite.db.First(&deck, "id = ?", "test_deck_id")
	cardsCount := deck.Remaining
	assert.Equal(t, cardsCount, 4)

	drawCount := 3
	resp, err := http.Get(testSuite.ts.URL + "/deck/test_deck_id/draw?count=" + strconv.Itoa(drawCount))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	testSuite.db.First(&deck, "id = ?", "test_deck_id")
	assert.Equal(t, cardsCount-drawCount, deck.Remaining)

	drawnCard := []model.Card{}
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&drawnCard))
	assert.Equal(t, drawnCard, mockDeck.Cards[len(mockDeck.Cards)-drawCount:])
}

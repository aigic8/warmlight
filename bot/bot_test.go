package bot

import (
	"testing"

	"github.com/aigic8/warmlight/db"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
)

const TEST_DB_URL = "file::memory:"

type reactNewUserTestCase struct {
	Name      string
	UserText  string
	ReplyText string
}

func TestReactNewUser(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{DB: DB}

	testCases := []reactNewUserTestCase{
		{Name: "normal", UserText: "/start", ReplyText: strWelcomeToBot(firstName)},
		{Name: "lostData", UserText: "bla", ReplyText: strYourDataIsLost(firstName)},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			update := makeTestMessageUpdate(int64(userID), firstName, tc.UserText)
			r, err := h.reactNewUser(user, update)
			assert.Nil(t, err)
			assert.Equal(t, len(r.Messages), 1)
			assert.Equal(t, r.Messages[0].Text, tc.ReplyText)
		})
	}
}

func TestReactStateNormal(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{DB: DB}

	updateText := "People who do crazy things are not necessarily crazy\nsources: The social animal, Elliot Aronson\n#sociology #psychology"
	update := makeTestMessageUpdate(int64(userID), firstName, updateText)

	r, err := h.reactStateNormal(user, update)
	assert.Nil(t, err)
	assert.Equal(t, len(r.Messages), 1)
	assert.Equal(t, r.Messages[0].Text, strQuoteAdded)
}

func makeTestMessageUpdate(userID int64, firstName string, text string) *models.Update {
	return &models.Update{
		Message: &models.Message{
			ID:   1,
			Text: text,
			From: &models.User{
				ID:        userID,
				FirstName: firstName,
				IsBot:     false,
				IsPremium: false,
			},
		},
	}
}

type reactSetActiveSourceTestCase struct {
	Name  string
	Text  string
	Reply string
}

func TestReactSetActiveSource(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if _, err := DB.CreateSource(userID, "The social animal"); err != nil {
		panic(err)
	}

	h := Handlers{DB: DB}
	testCases := []reactSetActiveSourceTestCase{
		{Name: "normal", Text: "/setActiveSource The social animal, 20", Reply: strActiveSourceIsSet("The social animal", 20)},
		{Name: "withoutTimeout", Text: "/setActiveSource The social animal", Reply: strActiveSourceIsSet("The social animal", DEFAULT_ACTIVE_SOURCE_TIMEOUT)},
		{Name: "malformed", Text: "/setActiveSource The, social, animal", Reply: strMalformedSetActiveSource},
		{Name: "empty", Text: "/setActiveSource", Reply: strMalformedSetActiveSource},
		{Name: "sourceDoesNotExist", Text: "/setActiveSource Elliot Aronson", Reply: strSourceDoesExist("Elliot Aronson")},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			update := makeTestMessageUpdate(int64(userID), firstName, tc.Text)
			r, err := h.reactSetActiveSource(user, update)

			assert.Nil(t, err)
			assert.Equal(t, 1, len(r.Messages))
			assert.Equal(t, tc.Reply, r.Messages[0].Text)
		})
	}

}

func mustInitDB(URL string) *db.DB {
	DB, err := db.NewDB(URL)
	if err != nil {
		panic(err)
	}
	if err := DB.Init(); err != nil {
		panic(err)
	}
	return DB
}

package bot

import (
	"testing"

	"github.com/aigic8/warmlight/bot/utils"
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
	var userID uint64 = 1234
	var chatID uint64 = 1
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{DB: DB}

	testCases := []reactNewUserTestCase{
		{Name: "normal", UserText: COMMAND_START, ReplyText: strWelcomeToBot(firstName)},
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

func TestReactDefaultNormal(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint64 = 1234
	var chatID uint64 = 1
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{DB: DB}

	updateText := "People who do crazy things are not necessarily crazy\nsources: The social animal, Elliot Aronson\n#sociology #psychology"
	update := makeTestMessageUpdate(int64(userID), firstName, updateText)

	r, err := h.reactDefault(user, update)
	assert.Nil(t, err)
	assert.Equal(t, len(r.Messages), 1)
	assert.Equal(t, r.Messages[0].Text, strQuoteAdded)
}

func TestReactDefaultWithOutput(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint64 = 1234
	var userChatID uint64 = 1
	var outputChatID uint64 = 2
	firstName := "aigic8"
	user, _, err := DB.GetOrCreateUser(userID, userChatID, firstName)
	if err != nil {
		panic(err)
	}

	_, createdOuput, err := DB.GetOrCreateOutput(userID, outputChatID, "salam")
	if err != nil {
		panic(err)
	}

	if !createdOuput {
		panic(err)
	}

	h := Handlers{DB: DB}

	updateText := "People who do crazy things are not necessarily crazy\nsources: The social animal, Elliot Aronson\n#sociology #psychology"
	update := makeTestMessageUpdate(int64(userID), firstName, updateText)
	update.Message.Chat.ID = int64(userChatID)

	qouteObj, err := utils.ParseQuote(updateText)
	if err != nil {
		panic(err)
	}

	r, err := h.reactDefault(user, update)
	assert.Nil(t, err)
	assert.Equal(t, len(r.Messages), 2)
	if r.Messages[0].ChatID.(int64) == int64(userChatID) {
		assert.Equal(t, r.Messages[0].Text, strQuoteAdded)
		assert.Equal(t, r.Messages[1].Text, strQuote(qouteObj))
	} else {
		assert.Equal(t, r.Messages[1].Text, strQuoteAdded)
		assert.Equal(t, r.Messages[0].Text, strQuote(qouteObj))
	}
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
	var userID uint64 = 1234
	var chatID uint64 = 1
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
		{Name: "normal", Text: COMMAND_SET_ACTIVE_SOURCE + " The social animal, 20", Reply: strActiveSourceIsSet("The social animal", 20)},
		{Name: "withoutTimeout", Text: COMMAND_SET_ACTIVE_SOURCE + " The social animal", Reply: strActiveSourceIsSet("The social animal", DEFAULT_ACTIVE_SOURCE_TIMEOUT)},
		{Name: "malformed", Text: COMMAND_SET_ACTIVE_SOURCE + " The, social, animal", Reply: strMalformedSetActiveSource},
		{Name: "empty", Text: COMMAND_SET_ACTIVE_SOURCE, Reply: strMalformedSetActiveSource},
		{Name: "sourceDoesNotExist", Text: COMMAND_SET_ACTIVE_SOURCE + " Elliot Aronson", Reply: strSourceDoesExist("Elliot Aronson")},
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

type reactAddOutputTestCase struct {
	Name  string
	Text  string
	Reply string
}

func TestReactAddOutput(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint64 = 1234
	var userChatID uint64 = 1
	firstName := "aigic8"
	var outputChatID uint64 = 10
	outputChatTitle := "My quotes"

	user, _, err := DB.GetOrCreateUser(userID, userChatID, firstName)
	if err != nil {
		panic(err)
	}

	_, created, err := DB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created the first time")
	}

	addOuputSpace := COMMAND_ADD_OUTPUT + " "
	testCases := []reactAddOutputTestCase{
		{Name: "normal", Text: addOuputSpace + outputChatTitle, Reply: strOutputIsSet(outputChatTitle)},
		{Name: "alreadyActive", Text: addOuputSpace + outputChatTitle, Reply: strOutputIsAlreadyActive(outputChatTitle)},
		{Name: "notExist", Text: addOuputSpace + "I do not exist", Reply: strOutputNotFound("I do not exist")},
	}

	h := Handlers{DB: DB}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			update := makeTestMessageUpdate(int64(userID), firstName, tc.Text)
			r, err := h.reactAddOutput(user, update)

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

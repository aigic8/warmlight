package bot

import (
	"testing"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
)

const TEST_DB_URL = "postgresql://postgres:postgres@localhost:1616/warmlight_test"
const TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS = 60
const DB_TIMEOUT = 5 * time.Second

type reactNewUserTestCase struct {
	Name      string
	UserText  string
	ReplyText string
}

func TestReactNewUser(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB}

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
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB}

	updateText := "People who do crazy things are not necessarily crazy\nsources: The social animal, Elliot Aronson\n#sociology #psychology"
	update := makeTestMessageUpdate(int64(userID), firstName, updateText)

	r, err := h.reactDefault(user, update)
	assert.Nil(t, err)
	assert.Equal(t, len(r.Messages), 1)
	assert.Equal(t, r.Messages[0].Text, strQuoteAdded)
}

func TestReactDefaultWithOutput(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 2
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, userChatID, firstName)
	if err != nil {
		panic(err)
	}

	_, createdOutput, err := appDB.GetOrCreateOutput(userID, outputChatID, "salam")
	if err != nil {
		panic(err)
	}

	if !createdOutput {
		panic(err)
	}

	h := Handlers{db: appDB}

	updateText := "People who do crazy things are not necessarily crazy\nsources: The social animal, Elliot Aronson\n#sociology #sociology"
	update := makeTestMessageUpdate(int64(userID), firstName, updateText)
	update.Message.Chat.ID = int64(userChatID)

	quoteObj, err := utils.ParseQuote(updateText)
	if err != nil {
		panic(err)
	}

	r, err := h.reactDefault(user, update)
	assert.Nil(t, err)
	assert.Equal(t, len(r.Messages), 2)
	if r.Messages[0].ChatID.(int64) == int64(userChatID) {
		assert.Equal(t, r.Messages[0].Text, strQuoteAdded)
		assert.Equal(t, r.Messages[1].Text, strQuote(quoteObj))
	} else {
		assert.Equal(t, r.Messages[1].Text, strQuoteAdded)
		assert.Equal(t, r.Messages[0].Text, strQuote(quoteObj))
	}
}

func TestReactDeactivateSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	sourceName := "The social animal"
	_, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if _, err := appDB.CreateSource(userID, sourceName); err != nil {
		panic(err)
	}

	user, happened, err := appDB.SetActiveSource(userID, sourceName, time.Now().Add(100*time.Minute))
	if !happened {
		panic("happened should be true")
	}
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB, defaultActiveSourceTimeoutMins: TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS}

	r1, err := h.reactDeactivateSource(user, makeTestMessageUpdate(userID, firstName, COMMAND_DEACTIVATE_SOURCE))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r1.Messages))
	assert.Equal(t, strActiveSourceDeactivated(sourceName), r1.Messages[0].Text)

	user, err = appDB.GetUser(userID)
	if err != nil {
		panic(err)
	}
	r2, err := h.reactDeactivateSource(user, makeTestMessageUpdate(userID, firstName, COMMAND_DEACTIVATE_SOURCE))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r2.Messages))
	assert.Equal(t, strNoActiveSource, r2.Messages[0].Text)
}

type reactSetActiveSourceTestCase struct {
	Name  string
	Text  string
	Reply string
}

func TestReactSetActiveSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if _, err := appDB.CreateSource(userID, "The social animal"); err != nil {
		panic(err)
	}

	h := Handlers{db: appDB, defaultActiveSourceTimeoutMins: TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS}
	testCases := []reactSetActiveSourceTestCase{
		{Name: "normal", Text: COMMAND_SET_ACTIVE_SOURCE + " The social animal, 20", Reply: strActiveSourceIsSet("The social animal", 20)},
		{Name: "withoutTimeout", Text: COMMAND_SET_ACTIVE_SOURCE + " The social animal", Reply: strActiveSourceIsSet("The social animal", TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS)},
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
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	firstName := "aigic8"
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	user, _, err := appDB.GetOrCreateUser(userID, userChatID, firstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created the first time")
	}

	addOutputSpace := COMMAND_ADD_OUTPUT + " "
	testCases := []reactAddOutputTestCase{
		{Name: "normal", Text: addOutputSpace + outputChatTitle, Reply: strOutputIsSet(outputChatTitle)},
		{Name: "alreadyActive", Text: addOutputSpace + outputChatTitle, Reply: strOutputIsAlreadyActive(outputChatTitle)},
		{Name: "notExist", Text: addOutputSpace + "I do not exist", Reply: strOutputNotFound("I do not exist")},
	}

	h := Handlers{db: appDB}
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

func mustInitDB(URL string) *db.DB {
	appDB, err := db.NewDB(URL, DB_TIMEOUT)
	if err != nil {
		panic(err)
	}

	if err := appDB.DEBUGCleanDB(); err != nil {
		panic(err)
	}

	if err := appDB.Init(); err != nil {
		panic(err)
	}

	return appDB
}

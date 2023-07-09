package bot

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot/strs"
	"github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot/models"
	"github.com/hako/durafmt"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

const TEST_DB_URL = "postgresql://postgres:postgres@localhost/warmlight_dev"
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
		{Name: "normal", UserText: strs.COMMAND_START, ReplyText: strs.WelcomeToBot(firstName)},
		{Name: "lostData", UserText: "bla", ReplyText: strs.YourDataIsLost(firstName)},
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

func TestReactDefault(t *testing.T) {
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
	assert.Equal(t, r.Messages[0].Text, strs.QuoteAdded)
}

func TestReactGetLibraryToken(t *testing.T) {
	// TODO: test the case when user is the owner of the library
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	u, _, err := appDB.GetOrCreateUser(1, 1, "aigic8")
	if err != nil {
		panic(err)
	}

	UUIDLifetime := 30 * time.Minute
	h := Handlers{db: appDB, LibraryUUIDLifetime: UUIDLifetime}

	up1 := makeTestMessageUpdate(u.ID, u.FirstName, strs.COMMAND_GET_LIBRARY_TOKEN)
	r, err := h.reactGetLibraryToken(u, up1)
	assert.Nil(t, err)

	library, err := h.db.GetLibrary(u.LibraryID)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(r.Messages))
	UUIDLifetimeStr := durafmt.Parse(UUIDLifetime).String()
	expectedText1 := strs.YourLibraryToken(library.Token.UUID.String(), UUIDLifetimeStr)
	assert.Equal(t, expectedText1, r.Messages[0].Text)
}

func TestReactSetLibraryToken(t *testing.T) {
	// TODO: test if the token is expired
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	u1, _, err := appDB.GetOrCreateUser(1, 1, "aigic8")
	if err != nil {
		panic(err)
	}

	UUIDLifetime := 30 * time.Minute
	h := Handlers{db: appDB, LibraryUUIDLifetime: UUIDLifetime}

	up1 := makeTestMessageUpdate(u1.ID, u1.FirstName, strs.COMMAND_GET_LIBRARY_TOKEN)
	_, err = h.reactGetLibraryToken(u1, up1)
	if err != nil {
		panic(err)
	}

	u1Library, err := appDB.GetLibrary(u1.LibraryID)

	u2, _, err := appDB.GetOrCreateUser(2, 12, "aigic88")
	if err != nil {
		panic(err)
	}

	up2 := makeTestMessageUpdate(u2.ID, u2.FirstName, strs.COMMAND_GET_LIBRARY_TOKEN+" "+u1Library.Token.UUID.String())
	r2, err := h.reactSetLibraryToken(u2, up2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r2.Messages))
	assert.Equal(t, strs.MergeOrDeleteCurrentLibraryData, r2.Messages[0].Text)
	assert.Equal(t, utils.MergeOrDeleteCurrentLibraryReplyMarkup, r2.Messages[0].ReplyMarkup)
}

func TestReactStateEditingSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	sourceName := "Practical Statistics for Data Scientists"
	source, err := appDB.CreateSource(user.LibraryID, sourceName)
	if err != nil {
		panic(err)
	}

	if _, err = appDB.SetUserStateEditingSource(userID, source.ID); err != nil {
		panic(err)
	}

	h := Handlers{db: appDB}

	sourceInfoURL := "https://www.oreilly.com/library/view/practical-statistics-for/9781492072935/"
	sourceAuthor := "Peter Bruce"
	updateMessage := fmt.Sprintf("%s: %s\n%s: %s\n%s: %s", strs.SOURCE_KIND, "book", strs.SOURCE_BOOK_AUTHOR, sourceAuthor, strs.SOURCE_BOOK_INFO_URL, sourceInfoURL)
	update := makeTestMessageUpdate(userID, firstName, updateMessage)

	user, err = appDB.GetUser(userID)
	if err != nil {
		panic(err)
	}
	r, err := h.reactStateEditingSource(user, update)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r.Messages))

	expectedSourceData := db.SourceBookData{
		Author:     sourceAuthor,
		LinkToInfo: sourceInfoURL,
	}
	expectedSourceDataBytes, err := json.Marshal(&expectedSourceData)
	if err != nil {
		panic(err)
	}

	expectedSourceDataJson := pgtype.JSON{Status: pgtype.Present, Bytes: expectedSourceDataBytes}
	expectedText, err := strs.UpdatedSource(&db.Source{Name: sourceName, Kind: db.SourceKindBook, Data: expectedSourceDataJson})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expectedText, r.Messages[0].Text)
}

func TestReactStateConfirmingLibraryChange(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	u1, _, err := appDB.GetOrCreateUser(1, 123, "aigic8")
	if err != nil {
		panic(err)
	}

	u2, _, err := appDB.GetOrCreateUser(2, 321, "aigic2")
	if err != nil {
		panic(err)
	}

	u2, err = appDB.SetUserStateConfirmingLibraryChange(u2.ID, u1.LibraryID, db.ChangeLibraryDeleteMode)
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB}
	update := makeTestMessageUpdate(u2.ID, u2.FirstName, strs.ConfirmLibraryChangeYesAnswer)
	r, err := h.reactStateConfirmingLibraryChange(u2, update)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r.Messages))
	assert.Equal(t, strs.LibraryChangedSuccessfully, r.Messages[0].Text)
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
		assert.Equal(t, r.Messages[0].Text, strs.QuoteAdded)
		assert.Equal(t, r.Messages[1].Text, strs.Quote(quoteObj))
	} else {
		assert.Equal(t, r.Messages[1].Text, strs.QuoteAdded)
		assert.Equal(t, r.Messages[0].Text, strs.Quote(quoteObj))
	}
}

func TestReactDeactivateSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	sourceName := "The social animal"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if _, err := appDB.CreateSource(user.LibraryID, sourceName); err != nil {
		panic(err)
	}

	user, err = appDB.SetActiveSource(userID, sourceName, time.Now().Add(100*time.Minute))
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB, defaultActiveSourceTimeoutMins: TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS}

	r1, err := h.reactDeactivateSource(user, makeTestMessageUpdate(userID, firstName, strs.COMMAND_DEACTIVATE_SOURCE))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r1.Messages))
	assert.Equal(t, strs.ActiveSourceDeactivated(sourceName), r1.Messages[0].Text)

	user, err = appDB.GetUser(userID)
	if err != nil {
		panic(err)
	}
	r2, err := h.reactDeactivateSource(user, makeTestMessageUpdate(userID, firstName, strs.COMMAND_DEACTIVATE_SOURCE))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r2.Messages))
	assert.Equal(t, strs.NoActiveSource, r2.Messages[0].Text)
}

type reactSetActiveSourceTestCase struct {
	Name  string
	Text  string
	Reply string
}

type reactGetSourcesTestCase struct {
	Name    string
	Query   string
	Results []string
}

func TestReactGetSources(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	user, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	s1Name := "aaa"
	s2Name := "aaaa"
	s3Name := "aaaaa"
	availableSources := []string{s1Name, s2Name, s3Name}
	sourcesMap := map[string]db.Source{}
	for _, sourceName := range availableSources {
		source, err := appDB.CreateSource(user.LibraryID, sourceName)
		if err != nil {
			panic(err)
		}
		sourcesMap[sourceName] = *source
	}

	s1New, err := appDB.SetSourceBook(user.LibraryID, sourcesMap[s1Name].ID, nil)
	if err != nil {
		panic(err)
	}
	sourcesMap[s1Name] = *s1New

	h := Handlers{db: appDB}

	testCases := []reactGetSourcesTestCase{
		{Name: "normal", Query: s1Name, Results: []string{s1Name, s2Name, s3Name}},
		{Name: "noResult", Query: "bbb", Results: []string{}},
		{Name: "withSourceKindFilter", Query: s1Name + " @unknown", Results: []string{s2Name, s3Name}},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			update := makeTestMessageUpdate(userID, firstName, strs.COMMAND_GET_SOURCES+" "+tc.Query)
			r, err := h.reactGetSources(user, update)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(r.Messages))

			sources := []db.Source{}
			for _, sourceName := range tc.Results {
				sources = append(sources, sourcesMap[sourceName])
			}

			assert.Equal(t, strs.ListOfSources(sources), r.Messages[0].Text)
			assert.Equal(t, utils.SourcesReplyMarkup(sources, true, true), r.Messages[0].ReplyMarkup)
		})
	}
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

	if _, err := appDB.CreateSource(user.LibraryID, "The social animal"); err != nil {
		panic(err)
	}

	h := Handlers{db: appDB, defaultActiveSourceTimeoutMins: TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS}
	testCases := []reactSetActiveSourceTestCase{
		{Name: "normal", Text: strs.COMMAND_SET_ACTIVE_SOURCE + " The social animal, 20", Reply: strs.ActiveSourceIsSet("The social animal", 20)},
		{Name: "withoutTimeout", Text: strs.COMMAND_SET_ACTIVE_SOURCE + " The social animal", Reply: strs.ActiveSourceIsSet("The social animal", TEST_DEFAULT_ACTIVE_SOURCE_TIMEOUT_MINS)},
		{Name: "malformed", Text: strs.COMMAND_SET_ACTIVE_SOURCE + " The, social, animal", Reply: strs.MalformedSetActiveSource},
		{Name: "empty", Text: strs.COMMAND_SET_ACTIVE_SOURCE, Reply: strs.MalformedSetActiveSource},
		{Name: "sourceDoesNotExist", Text: strs.COMMAND_SET_ACTIVE_SOURCE + " Elliot Aronson", Reply: strs.SourceDoesNotExist("Elliot Aronson")},
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

// type reactAddOutputTestCase struct {
// 	Name  string
// 	Text  string
// 	Reply string
// }

// func TestReactAddOutput(t *testing.T) {
// 	appDB := mustInitDB(TEST_DB_URL)
// 	defer appDB.Close()
// 	var userID int64 = 1234
// 	var userChatID int64 = 1
// 	firstName := "aigic8"
// 	var outputChatID int64 = 10
// 	outputChatTitle := "My quotes"

// 	user, _, err := appDB.GetOrCreateUser(userID, userChatID, firstName)
// 	if err != nil {
// 		panic(err)
// 	}

// 	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
// 	if err != nil {
// 		panic(err)
// 	}

// 	if created == false {
// 		panic("output should be created the first time")
// 	}

// 	addOutputSpace := COMMAND_ADD_OUTPUT + " "
// 	testCases := []reactAddOutputTestCase{
// 		{Name: "normal", Text: addOutputSpace + outputChatTitle, Reply: strOutputIsSet(outputChatTitle)},
// 		{Name: "alreadyActive", Text: addOutputSpace + outputChatTitle, Reply: strOutputIsAlreadyActive(outputChatTitle)},
// 		{Name: "notExist", Text: addOutputSpace + "I do not exist", Reply: strOutputNotFound("I do not exist")},
// 	}

// 	h := Handlers{db: appDB}
// 	for _, tc := range testCases {
// 		t.Run(tc.Name, func(t *testing.T) {
// 			update := makeTestMessageUpdate(int64(userID), firstName, tc.Text)
// 			r, err := h.reactAddOutput(user, update)

// 			assert.Nil(t, err)
// 			assert.Equal(t, 1, len(r.Messages))
// 			assert.Equal(t, tc.Reply, r.Messages[0].Text)
// 		})
// 	}

// }

func TestReactGetOutputs(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	firstName := "aigic8"

	var output1ChatID int64 = 10
	output1Title := "My quotes"
	var output2ChatID int64 = 11
	output2Title := "Best quotes"

	user, _, err := appDB.GetOrCreateUser(userID, userChatID, firstName)
	if err != nil {
		panic(err)
	}

	h := Handlers{db: appDB}
	r1, err := h.reactGetOutputs(user, makeTestMessageUpdate(userID, firstName, strs.COMMAND_GET_OUTPUTS))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r1.Messages))
	assert.Equal(t, strs.ListOfYourOutputs([]db.Output{}), r1.Messages[0].Text)

	_, created, err := appDB.GetOrCreateOutput(userID, output1ChatID, output1Title)
	if err != nil {
		panic(err)
	}
	if created == false {
		panic("output should be created the first time")
	}

	output1, err := appDB.ActivateOutput(userID, output1ChatID)
	if err != nil {
		panic(err)
	}

	output2, created, err := appDB.GetOrCreateOutput(userID, output2ChatID, output2Title)
	if err != nil {
		panic(err)
	}
	if created == false {
		panic("output should be created the first time")
	}

	user, err = appDB.GetUser(userID)
	if err != nil {
		panic(err)
	}

	r2, err := h.reactGetOutputs(user, makeTestMessageUpdate(userID, firstName, strs.COMMAND_GET_OUTPUTS))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r2.Messages))
	msgText := r2.Messages[0].Text

	isTextValid := msgText == strs.ListOfYourOutputs([]db.Output{*output1, *output2}) ||
		msgText == strs.ListOfYourOutputs([]db.Output{*output2, *output1})
	assert.True(t, isTextValid, "text is not valid:\n%s", msgText)

	output1CallbackData, err := utils.MakeToggleOutputStateCallback(output1)
	if err != nil {
		panic(err)
	}

	output2CallbackData, err := utils.MakeToggleOutputStateCallback(output2)
	if err != nil {
		panic(err)
	}

	btn1Text := output1Title + " - " + "active"
	btn2Text := output2Title + " - " + "deactive"
	expectedButtons := map[string]models.InlineKeyboardButton{
		btn1Text: {Text: btn1Text, CallbackData: output1CallbackData},
		btn2Text: {Text: btn2Text, CallbackData: output2CallbackData},
	}

	foundButtons := map[string]bool{
		btn1Text: false,
		btn2Text: false,
	}

	keyboard := r2.Messages[0].ReplyMarkup.(models.InlineKeyboardMarkup).InlineKeyboard
	assert.Equal(t, len(expectedButtons), len(keyboard))

	for _, row := range keyboard {
		btnText := row[0].Text
		btn, btnFound := expectedButtons[btnText]
		assert.True(t, btnFound, "button with title '%s' was not expected", btnText)
		assert.Equal(t, btn.CallbackData, row[0].CallbackData)
		foundButtons[btnText] = true
	}

	for text, found := range foundButtons {
		assert.True(t, found, "button with text '%s' was not found", text)
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

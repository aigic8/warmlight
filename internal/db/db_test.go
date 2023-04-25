package db

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/aigic8/warmlight/internal/db/base"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

const TEST_DB_URL = "postgresql://postgres:postgres@localhost:1616/warmlight_test"
const DB_TIMEOUT = 2 * time.Second

// TODO refactor to standard testing methods

func TestNewDB(t *testing.T) {
	appDB, err := NewDB(TEST_DB_URL, DB_TIMEOUT)
	assert.Nil(t, err)
	appDB.Close()
}

func TestDBInit(t *testing.T) {
	appDB, err := NewDB(TEST_DB_URL, DB_TIMEOUT)
	if err != nil {
		panic(err)
	}
	defer appDB.Close()

	assert.Nil(t, appDB.Init())
}

func TestDBGetOrCreateUser(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var ID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"

	user, isCreated, err := appDB.GetOrCreateUser(ID, chatID, firstName)
	assert.Nil(t, err)
	assert.True(t, isCreated)
	assert.Equal(t, ID, user.ID)
	assert.Equal(t, UserStateNormal, user.State)
	assert.Equal(t, user.FirstName, firstName)

	user2, isCreated2, err := appDB.GetOrCreateUser(ID, chatID, firstName)
	assert.Nil(t, err)
	assert.False(t, isCreated2)
	assert.Equal(t, ID, user2.ID)
}

func TestDBSetUserStateNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var ID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"

	_, created, err := appDB.GetOrCreateUser(ID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if !created {
		panic("user should be created")
	}

	user, err := appDB.SetUserStateEditingSource(ID, 1)
	if err != nil {
		panic(err)
	}

	if user.State != UserStateEditingSource {
		panic("user state should be editing source")
	}

	resUser, err := appDB.SetUserStateNormal(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, UserStateNormal, resUser.State)
	assert.Equal(t, pgtype.Null, resUser.StateData.Status)
}

func TestDBSetUserStateEditingSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var ID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"
	var sourceID int64 = 10

	_, created, err := appDB.GetOrCreateUser(ID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	if !created {
		panic("user should be created")
	}

	user, err := appDB.SetUserStateEditingSource(ID, sourceID)
	assert.Nil(t, err)
	assert.Equal(t, UserStateEditingSource, user.State)
	assert.Equal(t, pgtype.Present, user.StateData.Status)

	var stateData StateEditingSourceData
	if err := json.Unmarshal(user.StateData.Bytes, &stateData); err != nil {
		panic(err)
	}
	assert.Equal(t, sourceID, stateData.SourceID)
}

func TestDBCreateSource(t *testing.T) {
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

	source, err := appDB.CreateSource(user.ID, sourceName)
	assert.Nil(t, err)
	assert.Equal(t, sourceName, source.Name)
	assert.Equal(t, userID, source.UserID)
}

func TestDBCreateSourceAlreadyExist(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.CreateSource(user.ID, sourceName)
	assert.NotNil(t, err)
}

func TestDBGetSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	source, err := appDB.GetSource(user.ID, sourceName)
	assert.Nil(t, err)
	assert.Equal(t, sourceName, source.Name)
}

type setSourceBookTestCase struct {
	Name       string
	SourceID   int64
	SourceName string
	SourceData *SourceBookData
}

func TestDBSetSourceBook(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 10
	var chatID int64 = 1
	firstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	socialAnimalName := "The social animal"
	tyrannyOfMeritName := "The tyranny of merit"
	availableSources := []string{socialAnimalName, tyrannyOfMeritName}
	sourceIDs := map[string]int64{}
	for _, sourceName := range availableSources {
		source, err := appDB.CreateSource(userID, sourceName)
		if err != nil {
			panic(err)
		}
		sourceIDs[sourceName] = source.ID
	}

	socialAnimalSD := SourceBookData{Author: "Elliot Aronson", LinkToInfo: "https://wikipedia.com/the-social-animal", LinkToAuthor: "https://wikipedia.com/elliot-aronson"}
	testCases := []setSourceBookTestCase{
		{Name: "normal", SourceID: sourceIDs[socialAnimalName], SourceName: socialAnimalName, SourceData: &socialAnimalSD},
		{Name: "nil", SourceID: sourceIDs[tyrannyOfMeritName], SourceName: tyrannyOfMeritName},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err = appDB.SetSourceBook(userID, tc.SourceID, tc.SourceData)
			assert.Nil(t, err)

			source, err := appDB.GetSource(userID, tc.SourceName)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, base.SourceKindBook, source.Kind)
			if tc.SourceData != nil {
				assert.Equal(t, pgtype.Present, source.Data.Status)

				var resSourceData SourceBookData
				if err = json.Unmarshal(source.Data.Bytes, &resSourceData); err != nil {
					panic(err)
				}

				assert.Equal(t, tc.SourceData.Author, resSourceData.Author)
				assert.Equal(t, tc.SourceData.LinkToAuthor, resSourceData.LinkToAuthor)
				assert.Equal(t, tc.SourceData.LinkToInfo, resSourceData.LinkToInfo)
			} else {
				assert.Equal(t, pgtype.Null, source.Data.Status)
			}
		})
	}
}

type setSourceUnknownTestCase struct {
	Name       string
	SourceID   int64
	SourceName string
}

func TestDBSetSourceUnknown(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 10
	var chatID int64 = 1
	firstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	socialAnimalName := "The social animal"
	availableSources := []string{socialAnimalName, "The tyranny of merit"}
	sourceIDs := map[string]int64{}
	for _, sourceName := range availableSources {
		source, err := appDB.CreateSource(userID, sourceName)
		if err != nil {
			panic(err)
		}

		// since default source kind is "unknown" we need to change it
		// to book before testing to make sure SetSourceUnknown works
		_, err = appDB.SetSourceBook(userID, source.ID, nil)
		if err != nil {
			panic(err)
		}

		sourceIDs[sourceName] = source.ID
	}

	testCases := []setSourceUnknownTestCase{
		{Name: "normal", SourceID: sourceIDs[socialAnimalName], SourceName: socialAnimalName},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			_, err = appDB.SetSourceUnknown(userID, tc.SourceID)
			assert.Nil(t, err)

			source, err := appDB.GetSource(userID, tc.SourceName)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, base.SourceKindUnknown, source.Kind)
			assert.Equal(t, pgtype.Null, source.Data.Status)
		})
	}
}

type querySourcesTestCase struct {
	Name       string
	Query      string
	SourceKind string
	Limit      int32
	BaseID     int64
	Before     bool
	Results    []string
}

func TestDBQuerySources(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 10
	var chatID int64 = 1
	firstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	articleSourceName := "aaa"
	availableSources := []string{articleSourceName, "aaaaa", "aaaaaaa", "aaaaaaaaaa"}
	sourceIDs := map[string]int64{}
	for _, sourceName := range availableSources {
		source, err := appDB.CreateSource(userID, sourceName)
		if err != nil {
			panic(err)
		}

		sourceIDs[sourceName] = source.ID
	}

	_, err = appDB.SetSourceArticle(userID, sourceIDs[articleSourceName], nil)
	if err != nil {
		panic(err)
	}

	testCases := []querySourcesTestCase{
		{Name: "normal", Query: "aaa", SourceKind: "", Results: availableSources, Limit: 10},
		{Name: "withLimit", Query: "aaa", SourceKind: "", Results: availableSources[:3], Limit: 3},
		{Name: "withSourceKind", Query: "aaa", SourceKind: "article", Results: []string{articleSourceName}, Limit: 10},
		{Name: "afterID", Query: "aaa", SourceKind: "", Results: availableSources[1:], Limit: 10, BaseID: sourceIDs["aaa"]},
		{Name: "beforeID", Query: "aaa", SourceKind: "", Results: []string{articleSourceName}, Limit: 10, BaseID: sourceIDs["aaaaa"], Before: true},
		{Name: "empty", Query: "bbb", SourceKind: "", Results: []string{}, Limit: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			sources, err := appDB.QuerySources(QuerySourcesParams{UserID: userID, NameQuery: tc.Query, SourceKind: tc.SourceKind, Limit: tc.Limit, BaseID: tc.BaseID, Before: tc.Before})
			assert.Nil(t, err)

			sourceNames := []string{}
			for _, source := range sources {
				sourceNames = append(sourceNames, source.Name)
			}
			assert.ElementsMatch(t, tc.Results, sourceNames)
		})
	}
}

func TestDBGetSourceErrDoesNotFound(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.GetSource(user.ID, sourceName)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDBSetActiveSourceNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	userFirstName := "aigic8"
	var chatID int64 = 1
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * 5)

	user, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.SetActiveSource(userID, sourceName, activeSourceExpire)
	assert.Nil(t, err)
}

func TestDBSetActiveSourceNotExist(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * 5)

	_, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.SetActiveSource(userID, sourceName, activeSourceExpire)
	assert.Nil(t, err)
}

func TestDBDeactivateExpiredSources(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * -5)

	_, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.SetActiveSource(userID, sourceName, activeSourceExpire)
	if err != nil {
		panic(err)
	}

	users, err := appDB.DeactivateExpiredSources()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, userFirstName, users[0].FirstName)
	assert.Equal(t, userID, users[0].ID)
	assert.Equal(t, chatID, users[0].ChatID)

	user, err := appDB.GetUser(userID)
	if err != nil {
		panic(err)
	}

	assert.False(t, user.ActiveSource.Valid)
	assert.False(t, user.ActiveSourceExpire.Valid)
}

func TestDBDeactivateSource(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var chatID int64 = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Hour * 5)

	_, _, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = appDB.SetActiveSource(userID, sourceName, activeSourceExpire)
	if err != nil {
		panic(err)
	}

	_, err = appDB.DeactivateSource(userID)
	assert.Nil(t, err)

	user, created, err := appDB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	if created {
		panic("user should not be created in the second time")
	}

	assert.False(t, user.ActiveSource.Valid)
	assert.False(t, user.ActiveSourceExpire.Valid)
}

func TestDBCreateQuoteWithData(t *testing.T) {
	var userID int64 = 1234
	var chatID int64 = 1
	text := "People who do crazy things are not necessarily crazy"
	mainSource := "The social animal"
	sources := []string{"The social animal", "Elliot Aronson"}
	tags := []string{"sociology", "psychology"}
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()

	_, _, err := appDB.GetOrCreateUser(userID, chatID, "aigic8")
	if err != nil {
		panic(err)
	}

	q, err := appDB.CreateQuoteWithData(userID, text, mainSource, tags, sources)

	assert.Nil(t, err)
	assert.Equal(t, q.Text, text)
	assert.Equal(t, q.MainSource, sql.NullString{Valid: true, String: mainSource})
}

func TestDBGetOrCreateOutputNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	output, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	assert.Nil(t, err)
	assert.True(t, created)
	assert.Equal(t, outputChatID, output.ChatID)
	assert.Equal(t, outputChatTitle, output.Title)
	assert.False(t, output.IsActive)
}

func TestDBGetOrCreateOutputAlreadyCreated(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created in the first time")
	}

	_, created2, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	assert.Nil(t, err)
	assert.False(t, created2)
}

func TestDBGetOutputNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created in the first time")
	}

	output, err := appDB.GetOutput(userID, outputChatID)
	assert.Nil(t, err)
	assert.Equal(t, outputChatID, output.ChatID)
	assert.Equal(t, outputChatTitle, output.Title)
}

func TestDBSetOutputActiveNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created in the first time")
	}

	_, err = appDB.ActivateOutput(userID, outputChatID)
	assert.Nil(t, err)

	output, err := appDB.GetOutput(userID, outputChatID)
	assert.Nil(t, err)
	assert.Equal(t, outputChatID, output.ChatID)
	assert.Equal(t, outputChatTitle, output.Title)
	assert.True(t, output.IsActive)
}

func TestDBDeleteOutputNormal(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created in the first time")
	}

	err = appDB.DeleteOutput(userID, outputChatID)
	assert.Nil(t, err)

	_, err = appDB.GetOutput(userID, outputChatID)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDBDeleteOutputNotExist(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	defer appDB.Close()
	var userID int64 = 1234
	var userChatID int64 = 1

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}
	err = appDB.DeleteOutput(userID, 10)
	assert.Nil(t, err)
}

func TestDBGetOutputs(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 1234
	var userChatID int64 = 1
	var outputChatID int64 = 10
	outputChatTitle := "My quotes"

	userFirstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, userChatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, created, err := appDB.GetOrCreateOutput(userID, outputChatID, outputChatTitle)
	if err != nil {
		panic(err)
	}

	if created == false {
		panic("output should be created in the first time")
	}

	outputs, err := appDB.GetOutputs(userID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(outputs))
	assert.Equal(t, outputChatID, outputs[0].ChatID)
	assert.Equal(t, outputChatTitle, outputs[0].Title)
}

type searchQuotesTestCase struct {
	Name    string
	Query   string
	Results []string
}

func TestDBSearchQuotes(t *testing.T) {
	appDB := mustInitDB(TEST_DB_URL)
	var userID int64 = 1234
	var chatID int64 = 1
	firstName := "aigic8"

	_, _, err := appDB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	q1Text := "people who do crazy things are not necessarily crazy"
	q2Text := "Premature optimization is the root of all evil"
	quotes := []struct {
		Text   string
		Source string
	}{
		{Text: q1Text, Source: "The social animal"},
		{Text: q2Text, Source: "Donald Knuth"},
	}

	for _, quote := range quotes {
		_, err := appDB.CreateQuoteWithData(userID, quote.Text, quote.Source, []string{}, []string{quote.Source})
		if err != nil {
			panic(err)
		}
	}

	testCases := []searchQuotesTestCase{
		{Name: "normal", Query: "people & crazy", Results: []string{q1Text}},
		{Name: "mainSource", Query: "Knuth", Results: []string{q2Text}},
		{Name: "empty", Query: "umbrella", Results: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := appDB.SearchQuotes(userID, tc.Query, 10)
			assert.Nil(t, err)

			resultTexts := []string{}
			for _, quote := range result {
				resultTexts = append(resultTexts, quote.Text)
			}

			assert.ElementsMatch(t, tc.Results, resultTexts)
		})
	}

}

func mustInitDB(URL string) *DB {
	appDB, err := NewDB(URL, DB_TIMEOUT)
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

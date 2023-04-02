package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const TEST_DB_URL = "file::memory:"

// TODO refactor to standard testing methods

func TestNewDB(t *testing.T) {
	db, err := NewDB(TEST_DB_URL)
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	assert.Nil(t, err)
}

func TestDBInit(t *testing.T) {
	db, err := NewDB(TEST_DB_URL)
	if err != nil {
		panic(err)
	}

	assert.Nil(t, db.Init())
}

func TestDBGetOrCreateUser(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var ID uint = 1234
	var chatID uint = 1
	firstName := "aigic8"

	user, isCreated, err := DB.GetOrCreateUser(ID, chatID, firstName)
	assert.Nil(t, err)
	assert.True(t, isCreated)
	assert.Equal(t, user.ID, ID)
	assert.Equal(t, user.FirstName, firstName)

	user2, isCreated2, err := DB.GetOrCreateUser(ID, chatID, firstName)
	assert.Nil(t, err)
	assert.False(t, isCreated2)
	assert.Equal(t, user2.ID, ID)
}

func TestDBCreateSource(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	firstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := DB.GetOrCreateUser(userID, chatID, firstName)
	if err != nil {
		panic(err)
	}

	source, err := DB.CreateSource(user.ID, sourceName)
	assert.Nil(t, err)
	assert.Equal(t, sourceName, source.Name)
	assert.Equal(t, userID, source.UserID)
}

func TestDBCreateSourceAlreadyExist(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = DB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	_, err = DB.CreateSource(user.ID, sourceName)
	assert.NotNil(t, err)
}

func TestDBGetSource(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = DB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	source, err := DB.GetSource(user.ID, sourceName)
	assert.Nil(t, err)
	assert.Equal(t, sourceName, source.Name)
}

func TestDBGetSourceErrDoesNotFound(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"

	user, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = DB.GetSource(user.ID, sourceName)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDBSetActiveSourceNormal(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	userFirstName := "aigic8"
	var chatID uint = 1
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * 5)

	user, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	_, err = DB.CreateSource(user.ID, sourceName)
	if err != nil {
		panic(err)
	}

	effected, err := DB.SetActiveSource(userID, sourceName, activeSourceExpire)
	assert.Nil(t, err)
	assert.True(t, effected)
}

func TestDBSetActiveSourceNotExist(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * 5)

	_, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	effected, err := DB.SetActiveSource(userID, sourceName, activeSourceExpire)
	assert.Nil(t, err)
	assert.True(t, effected)
}

func TestDBDeactivateExpiredSources(t *testing.T) {
	DB := mustInitDB(TEST_DB_URL)
	var userID uint = 1234
	var chatID uint = 1
	userFirstName := "aigic8"
	sourceName := "The social animal"
	activeSourceExpire := time.Now().Add(time.Minute * -5)

	_, _, err := DB.GetOrCreateUser(userID, chatID, userFirstName)
	if err != nil {
		panic(err)
	}

	effected, err := DB.SetActiveSource(userID, sourceName, activeSourceExpire)
	if err != nil {
		panic(err)
	}

	if !effected {
		panic("effected should be true")
	}

	users, err := DB.DeactivateExpiredSources()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, userFirstName, users[0].FirstName)
	assert.Equal(t, userID, users[0].ID)
	assert.Equal(t, chatID, users[0].ChatID)

	user, err := DB.GetUser(userID)
	if err != nil {
		panic(err)
	}

	assert.False(t, user.ActiveSource.Valid)
	assert.False(t, user.ActiveSourceExpire.Valid)
}

func TestDBCreateQuoteWithData(t *testing.T) {
	var userID uint = 1234
	var chatID uint = 1
	text := "People who do crazy things are not necessarily crazy"
	mainSource := "The social animal"
	sources := []string{"The social animal", "Elliot Aronson"}
	tags := []string{"sociology", "psychology"}
	db, err := NewDB(TEST_DB_URL)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := db.Init(); err != nil {
		panic(err)
	}

	_, _, err = db.GetOrCreateUser(userID, chatID, "aigic8")
	if err != nil {
		panic(err)
	}

	if err := db.debugClean(); err != nil {
		panic(err)
	}

	q, err := db.CreateQuoteWithData(userID, text, mainSource, tags, sources)

	sourceNames := make([]string, 0, len(q.Sources))
	for _, source := range q.Sources {
		sourceNames = append(sourceNames, source.Name)
	}

	tagNames := make([]string, 0, len(tags))
	for _, tag := range q.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	assert.Nil(t, err)
	assert.Equal(t, q.Text, text)
	assert.Equal(t, q.MainSource, &mainSource)
	assert.ElementsMatch(t, sourceNames, sources)
	assert.ElementsMatch(t, tagNames, tags)
}

func mustInitDB(URL string) *DB {
	DB, err := NewDB(URL)
	if err != nil {
		panic(err)
	}

	if err := DB.Init(); err != nil {
		panic(err)
	}

	return DB
}

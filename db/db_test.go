package db

import (
	"testing"

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
	firstName := "aigic8"

	user, isCreated, err := DB.GetOrCreateUser(ID, firstName)
	assert.Nil(t, err)
	assert.True(t, isCreated)
	assert.Equal(t, user.ID, ID)
	assert.Equal(t, user.FirstName, firstName)

	user2, isCreated2, err := DB.GetOrCreateUser(ID, firstName)
	assert.Nil(t, err)
	assert.False(t, isCreated2)
	assert.Equal(t, user2.ID, ID)
}

func TestDBCreateQuoteWithData(t *testing.T) {
	var userID uint = 1234
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

	_, _, err = db.GetOrCreateUser(userID, "aigic8")
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

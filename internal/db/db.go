package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/aigic8/warmlight/internal/db/base"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNotFound = pgx.ErrNoRows

type User = base.User
type UserState = base.UserState
type Output = base.Output
type Source = base.Source
type SourceKind = base.SourceKind
type QuoteSearchResult = base.SearchQuotesRow
type CreateQuoteResult = base.CreateQuoteRow
type Library = base.Library

const SourceKindUnknown = base.SourceKindUnknown
const SourceKindBook = base.SourceKindBook
const SourceKindPerson = base.SourceKindPerson
const SourceKindArticle = base.SourceKindArticle

const UserStateNormal = base.UserStateNormal
const UserStateEditingSource = base.UserStateEditingSource
const UserStateChangingLibrary = base.UserStateChangingLibrary
const UserStateConfirmingLibraryChange = base.UserStateConfirmingLibraryChange

const ChangeLibraryMergeMode = "merge"
const ChangeLibraryDeleteMode = "delete"

type DB struct {
	pool    *pgxpool.Pool
	q       *base.Queries
	Timeout time.Duration
}

var VALID_SOURCE_KINDS []string = []string{"unknown", "book", "person", "article"}

type (
	SourceBookData struct {
		Author       string `json:"author,omitempty"`
		LinkToInfo   string `json:"linkToInfo,omitempty"`
		LinkToAuthor string `json:"linkToAuthor,omitempty"`
	}

	SourceArticleData struct {
		URL    string `json:"url,omitempty"`
		Author string `json:"author,omitempty"`
	}

	// TODO: convert bornOn and deathOn to integers (only year is important)
	SourcePersonData struct {
		LinkToInfo string    `json:"linkToInfo,omitempty"`
		Title      string    `json:"title,omitempty"`
		BornOn     time.Time `json:"bornOn,omitempty"`
		DeathOn    time.Time `json:"deathOn,omitempty"`
	}
)

type (
	StateEditingSourceData struct {
		SourceID int64 `json:"sourceID"`
	}

	StateChangingLibraryData struct {
		LibraryID int64 `json:"libraryID"`
	}

	StateConfirmingLibraryChangeData struct {
		LibraryID int64  `json:"libraryID"`
		Mode      string `json:"mode"`
	}
)

func NewDB(URL string, timeout time.Duration) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	db, err := pgxpool.Connect(ctx, URL)
	if err != nil {
		return nil, err
	}

	q := base.New(db)

	return &DB{pool: db, q: q, Timeout: timeout}, nil
}

func (db *DB) Init() error {
	// TODO: create init function
	return nil
}

func (db *DB) GetUser(ID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.GetUser(ctx, ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) SetUserStateNormal(userID int64) (*User, error) {
	// TODO: test SetUserStateNormal
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.SetUserState(ctx, base.SetUserStateParams{ID: userID, State: UserStateNormal, StateData: pgtype.JSON{Status: pgtype.Null}})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) SetUserStateEditingSource(userID int64, sourceID int64) (*User, error) {
	// TODO: test SetUserStateEditingSource
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	data := StateEditingSourceData{SourceID: sourceID}
	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	stateData := pgtype.JSON{Bytes: dataBytes, Status: pgtype.Present}

	user, err := db.q.SetUserState(ctx, base.SetUserStateParams{ID: userID, State: UserStateEditingSource, StateData: stateData})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) SetUserStateChangingLibrary(userID int64, libraryID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	data := StateChangingLibraryData{LibraryID: libraryID}
	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	stateData := pgtype.JSON{Bytes: dataBytes, Status: pgtype.Present}

	user, err := db.q.SetUserState(ctx, base.SetUserStateParams{ID: userID, State: UserStateChangingLibrary, StateData: stateData})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) SetUserStateConfirmingLibraryChange(userID int64, libraryID int64, mode string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	data := StateConfirmingLibraryChangeData{LibraryID: libraryID, Mode: mode}
	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	stateData := pgtype.JSON{Bytes: dataBytes, Status: pgtype.Present}

	user, err := db.q.SetUserState(ctx, base.SetUserStateParams{ID: userID, State: UserStateConfirmingLibraryChange, StateData: stateData})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetOrCreateUser(ID, ChatID int64, firstName string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.GetUser(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c, err := db.pool.Acquire(context.Background())
			if err != nil {
				return nil, false, err
			}
			defer c.Release()

			tx, err := c.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				return nil, false, err
			}

			defer func() {
				if err != nil {
					tx.Rollback(context.Background())
				} else {
					tx.Commit(context.Background())
				}
			}()

			q := db.q.WithTx(tx)
			library, err := q.CreateLibrary(ctx, ID)
			if err != nil {
				return nil, false, err
			}

			user, err := q.CreateUser(ctx, base.CreateUserParams{ID: ID, ChatID: ChatID, FirstName: firstName, LibraryID: library.ID})
			if err != nil {
				return nil, false, err
			}
			return &user, true, nil
		}
		return nil, false, err
	}

	return &user, false, nil
}

func (db *DB) DeleteUserCurrentLibraryAndMigrateTo(userID, currLibraryID, newLibraryID int64) error {
	c, err := db.pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer c.Release()

	tx, err := c.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	q := db.q.WithTx(tx)

	ctx, cancel := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel()
	if err = q.DeleteQuotesTagsInLibrary(ctx, currLibraryID); err != nil {
		return err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel2()
	if err = q.DeleteQuotesSourcesInLibrary(ctx2, currLibraryID); err != nil {
		return err
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel3()
	if err = q.DeleteQuotesInLibrary(ctx3, currLibraryID); err != nil {
		return err
	}

	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel4()
	if err = q.DeleteTagsInLibrary(ctx4, currLibraryID); err != nil {
		return err
	}

	ctx5, cancel5 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel5()
	if err = q.DeleteSourcesInLibrary(ctx5, currLibraryID); err != nil {
		return err
	}

	ctx6, cancel6 := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel6()
	if _, err = q.SetUserLibrary(ctx6, base.SetUserLibraryParams{LibraryID: newLibraryID, ID: userID}); err != nil {
		return err
	}

	ctx7, cancel7 := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel7()
	if err = q.DeleteLibrary(ctx7, currLibraryID); err != nil {
		return err
	}

	return nil
}

func (db *DB) MergeUserCurrentLibraryAndMigrateTo(userID, currLibraryID, newLibraryID int64) error {
	c, err := db.pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer c.Release()

	tx, err := c.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	q := db.q.WithTx(tx)

	ctx, cancel := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel()
	if err = q.SetQuotesTagsLibrary(ctx, base.SetQuotesTagsLibraryParams{LibraryID: newLibraryID, LibraryID_2: currLibraryID}); err != nil {
		return err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel2()
	if err = q.SetQuotesSourcesLibrary(ctx2, base.SetQuotesSourcesLibraryParams{LibraryID: newLibraryID, LibraryID_2: currLibraryID}); err != nil {
		return err
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel3()
	if err = q.SetQuotesLibrary(ctx3, base.SetQuotesLibraryParams{LibraryID: newLibraryID, LibraryID_2: currLibraryID}); err != nil {
		return err
	}

	ctx4, cancel4 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel4()
	if err = q.SetTagsLibrary(ctx4, base.SetTagsLibraryParams{LibraryID: newLibraryID, LibraryID_2: currLibraryID}); err != nil {
		return err
	}

	ctx5, cancel5 := context.WithTimeout(context.Background(), 2*db.Timeout)
	defer cancel5()
	if err = q.SetSourcesLibrary(ctx5, base.SetSourcesLibraryParams{LibraryID: newLibraryID, LibraryID_2: currLibraryID}); err != nil {
		return err
	}

	ctx6, cancel6 := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel6()
	if _, err = q.SetUserLibrary(ctx6, base.SetUserLibraryParams{LibraryID: newLibraryID, ID: userID}); err != nil {
		return err
	}

	ctx7, cancel7 := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel7()
	if err = q.DeleteLibrary(ctx7, currLibraryID); err != nil {
		return err
	}

	return nil
}

func (db *DB) GetLibrary(libraryID int64) (*Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	library, err := db.q.GetLibrary(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	return &library, nil
}

func (db *DB) GetLibraryByUUID(UUID uuid.UUID) (*Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	library, err := db.q.GetLibraryByUUID(ctx, uuid.NullUUID{UUID: UUID, Valid: true})
	if err != nil {
		return nil, err
	}

	return &library, nil
}

func (db *DB) SetLibraryToken(libraryID int64, UUID uuid.UUID, expiresOn time.Time) (*Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	library, err := db.q.SetLibraryToken(ctx, base.SetLibraryTokenParams{
		ID:             libraryID,
		Token:          uuid.NullUUID{UUID: UUID, Valid: true},
		TokenExpiresOn: sql.NullTime{Time: expiresOn, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &library, nil
}

func (db *DB) DeleteLibraryToken(libraryID int64) (*Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	library, err := db.q.SetLibraryToken(ctx, base.SetLibraryTokenParams{
		ID:             libraryID,
		Token:          uuid.NullUUID{},
		TokenExpiresOn: sql.NullTime{},
	})
	if err != nil {
		return nil, err
	}
	return &library, nil
}

func (db *DB) CreateQuoteWithData(libraryID int64, text, mainSource string, tagNames []string, sourceNames []string) (*CreateQuoteResult, error) {
	c, err := db.pool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer c.Release()

	tx, err := c.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	q := db.q.WithTx(tx)

	mainSourceSql := sql.NullString{}
	if mainSource != "" {
		mainSourceSql = sql.NullString{Valid: true, String: mainSource}
	}
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	quote, err := q.CreateQuote(ctx, base.CreateQuoteParams{LibraryID: libraryID, Text: text, MainSource: mainSourceSql})
	if err != nil {
		return nil, err
	}

	for _, name := range tagNames {
		ctx, cancel := context.WithTimeout(context.Background(), db.Timeout*2)
		defer cancel()
		tagID, err := q.GetOrCreateTag(ctx, base.GetOrCreateTagParams{LibraryID: libraryID, Name: name})
		if err != nil {
			return nil, err
		}
		err = q.CreateQuotesTags(ctx, base.CreateQuotesTagsParams{Quote: quote.ID, Tag: tagID, LibraryID: libraryID})
		if err != nil {
			return nil, err
		}
	}

	for _, name := range sourceNames {
		ctx, cancel := context.WithTimeout(context.Background(), db.Timeout*2)
		defer cancel()
		sourceID, err := q.GetOrCreateSource(ctx, base.GetOrCreateSourceParams{LibraryID: libraryID, Name: name})
		if err != nil {
			return nil, err
		}
		err = q.CreateQuotesSources(ctx, base.CreateQuotesSourcesParams{Quote: quote.ID, Source: sourceID, LibraryID: libraryID})
		if err != nil {
			return nil, err
		}
	}

	return &quote, nil
}

func (db *DB) CreateSource(libraryID int64, name string) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.CreateSource(ctx, base.CreateSourceParams{LibraryID: libraryID, Name: name})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) GetSource(libraryID int64, name string) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.GetSource(ctx, base.GetSourceParams{LibraryID: libraryID, Name: name})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) GetSourceByID(libraryID int64, sourceID int64) (*Source, error) {
	// TODO: test GetSourceByID
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.GetSourceByID(ctx, base.GetSourceByIDParams{LibraryID: libraryID, ID: sourceID})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) SetSourceBook(libraryID int64, sourceID int64, sourceData *SourceBookData) (*Source, error) {
	var err error
	data := pgtype.JSON{Status: pgtype.Null}
	if sourceData != nil {
		data.Status = pgtype.Present
		data.Bytes, err = json.Marshal(sourceData)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.SetSourceData(ctx, base.SetSourceDataParams{LibraryID: libraryID, ID: sourceID, Kind: base.SourceKindBook, Data: data})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) SetSourceArticle(libraryID int64, sourceID int64, sourceData *SourceArticleData) (*Source, error) {
	var err error
	data := pgtype.JSON{Status: pgtype.Null}
	if sourceData != nil {
		data.Status = pgtype.Present
		data.Bytes, err = json.Marshal(sourceData)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.SetSourceData(ctx, base.SetSourceDataParams{LibraryID: libraryID, ID: sourceID, Kind: base.SourceKindArticle, Data: data})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) SetSourcePerson(libraryID int64, sourceID int64, sourceData *SourcePersonData) (*Source, error) {
	var err error
	data := pgtype.JSON{Status: pgtype.Null}
	if sourceData != nil {
		data.Status = pgtype.Present
		data.Bytes, err = json.Marshal(sourceData)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.SetSourceData(ctx, base.SetSourceDataParams{LibraryID: libraryID, ID: sourceID, Kind: base.SourceKindPerson, Data: data})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) SetSourceUnknown(libraryID int64, sourceID int64) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	data := pgtype.JSON{Status: pgtype.Null}
	source, err := db.q.SetSourceData(ctx, base.SetSourceDataParams{LibraryID: libraryID, ID: sourceID, Kind: base.SourceKindUnknown, Data: data})
	if err != nil {
		return nil, err
	}

	return &source, nil

}

func (db *DB) UpdateSource(libraryID int64, source *Source) (*Source, error) {
	if source == nil {
		return nil, errors.New("source is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	resSource, err := db.q.UpdateSource(ctx, base.UpdateSourceParams{Name: source.Name, Kind: source.Kind, Data: source.Data, ID: source.ID, LibraryID: libraryID})
	if err != nil {
		return nil, err
	}

	return &resSource, nil
}

type QuerySourcesParams struct {
	LibraryID  int64
	NameQuery  string
	SourceKind string
	Limit      int32
	BaseID     int64
	Before     bool
}

func (db *DB) QuerySources(p QuerySourcesParams) ([]Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	if p.Before {
		if p.SourceKind == "" {
			return db.q.QuerySourcesBefore(ctx, base.QuerySourcesBeforeParams{
				LibraryID: p.LibraryID,
				ID:        p.BaseID,
				Column3:   sql.NullString{Valid: true, String: p.NameQuery},
				Limit:     p.Limit,
			})
		} else {
			return db.q.QuerySourcesBeforeWithKind(ctx, base.QuerySourcesBeforeWithKindParams{
				LibraryID: p.LibraryID,
				ID:        p.BaseID,
				Kind:      base.SourceKind(p.SourceKind),
				Column4:   sql.NullString{Valid: true, String: p.NameQuery},
				Limit:     p.Limit,
			})
		}
	}

	if p.SourceKind == "" {
		return db.q.QuerySourcesAfter(ctx, base.QuerySourcesAfterParams{
			LibraryID: p.LibraryID,
			ID:        p.BaseID,
			Column3:   sql.NullString{Valid: true, String: p.NameQuery},
			Limit:     p.Limit,
		})
	}

	return db.q.QuerySourcesAfterWithKind(ctx, base.QuerySourcesAfterWithKindParams{
		LibraryID: p.LibraryID,
		ID:        p.BaseID,
		Kind:      base.SourceKind(p.SourceKind),
		Column4:   sql.NullString{Valid: true, String: p.NameQuery},
		Limit:     p.Limit,
	})
}

func (db *DB) SetActiveSource(userID int64, activeSourceStr string, activeSourceExpireTime time.Time) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.SetActiveSource(ctx, base.SetActiveSourceParams{
		ID:                 userID,
		ActiveSource:       sql.NullString{Valid: true, String: activeSourceStr},
		ActiveSourceExpire: sql.NullTime{Valid: true, Time: activeSourceExpireTime},
	})

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) DeactivateExpiredSources() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	return db.q.DeactivateExpiredSources(ctx)
}

func (db *DB) DeactivateSource(userID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.DeactivateSource(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetOutputs(userID int64) ([]Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	return db.q.GetOutputs(ctx, userID)
}

func (db *DB) GetOutput(userID int64, outputChatID int64) (*Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	output, err := db.q.GetOutput(ctx, base.GetOutputParams{UserID: userID, ChatID: outputChatID})
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (db *DB) ActivateOutput(userID int64, outputChatID int64) (*Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	output, err := db.q.ActivateOutput(ctx, base.ActivateOutputParams{UserID: userID, ChatID: outputChatID})
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (db *DB) DeactivateOutput(userID int64, outputChatID int64) (*Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	output, err := db.q.DeactivateOutput(ctx, base.DeactivateOutputParams{UserID: userID, ChatID: outputChatID})
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (db *DB) GetOrCreateOutput(userID int64, chatID int64, chatTitle string) (*Output, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	output, err := db.q.GetOutput(ctx, base.GetOutputParams{UserID: userID, ChatID: chatID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			output, err := db.q.CreateOutput(ctx, base.CreateOutputParams{UserID: userID, ChatID: chatID, Title: chatTitle})
			if err != nil {
				return nil, false, err
			}
			return &output, true, nil
		}

		return nil, false, err
	}

	return &output, false, nil
}

func (db *DB) SearchQuotes(libraryID int64, query string, limit int32) ([]QuoteSearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	return db.q.SearchQuotes(ctx, base.SearchQuotesParams{LibraryID: libraryID, ToTsquery: query, Limit: limit})
}

func (db *DB) DeleteOutput(userID int64, outputChatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	err := db.q.DeleteOutput(ctx, base.DeleteOutputParams{UserID: userID, ChatID: outputChatID})
	return err
}

func (db *DB) DEBUGCleanDB() error {
	if err := db.q.CleanOutputs(context.Background()); err != nil {
		return err
	}
	if err := db.q.CleanQuotesSources(context.Background()); err != nil {
		return err
	}

	if err := db.q.CleanQuotesTags(context.Background()); err != nil {
		return err
	}

	if err := db.q.CleanTags(context.Background()); err != nil {
		return err
	}

	if err := db.q.CleanSources(context.Background()); err != nil {
		return err
	}

	if err := db.q.CleanQuotes(context.TODO()); err != nil {
		return err
	}

	if err := db.q.CleanUsers(context.Background()); err != nil {
		return err
	}

	if err := db.q.CleanLibraries(context.Background()); err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() {
	db.pool.Close()
}

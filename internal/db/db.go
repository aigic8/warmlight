package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/aigic8/warmlight/internal/db/base"
	"github.com/aigic8/warmlight/pkg/bot/models"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNotFound = pgx.ErrNoRows

type User = base.User
type Output = base.Output
type Source = base.Source
type QuoteSearchResult = base.SearchQuotesRow
type CreateQuoteResult = base.CreateQuoteRow

type DB struct {
	pool    *pgxpool.Pool
	q       *base.Queries
	Timeout time.Duration
}

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
	// TODO create init function
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

func (db *DB) GetOrCreateUser(ID, ChatID int64, firstName string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	user, err := db.q.GetUser(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			user, err := db.q.CreateUser(ctx, base.CreateUserParams{ID: ID, ChatID: ChatID, FirstName: firstName})
			if err != nil {
				return nil, false, err
			}
			return &user, true, nil
		}
		return nil, false, err
	}

	return &user, false, nil
}

func (db *DB) CreateQuoteWithData(userID int64, text, mainSource string, tagNames []string, sourceNames []string) (*CreateQuoteResult, error) {
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
	quote, err := q.CreateQuote(ctx, base.CreateQuoteParams{UserID: userID, Text: text, MainSource: mainSourceSql})
	if err != nil {
		return nil, err
	}

	for _, name := range tagNames {
		ctx, cancel := context.WithTimeout(context.Background(), db.Timeout*2)
		defer cancel()
		tagID, err := q.GetOrCreateTag(ctx, base.GetOrCreateTagParams{UserID: userID, Name: name})
		if err != nil {
			return nil, err
		}
		err = q.CreateQuotesTags(ctx, base.CreateQuotesTagsParams{Quote: quote.ID, Tag: tagID})
		if err != nil {
			return nil, err
		}
	}

	for _, name := range sourceNames {
		ctx, cancel := context.WithTimeout(context.Background(), db.Timeout*2)
		defer cancel()
		sourceID, err := q.GetOrCreateSource(ctx, base.GetOrCreateSourceParams{UserID: userID, Name: name})
		if err != nil {
			return nil, err
		}
		err = q.CreateQuotesSources(ctx, base.CreateQuotesSourcesParams{Quote: quote.ID, Source: sourceID})
		if err != nil {
			return nil, err
		}
	}

	return &quote, nil
}

func (db *DB) CreateSource(userID int64, name string) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.CreateSource(ctx, base.CreateSourceParams{UserID: userID, Name: name})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) GetSource(userID int64, name string) (*Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	source, err := db.q.GetSource(ctx, base.GetSourceParams{UserID: userID, Name: name})
	if err != nil {
		return nil, err
	}

	return &source, nil
}

func (db *DB) SetActiveSource(userID int64, activeSourceStr string, activeSourceExpireTime time.Time) (*User, error) {
	// FIXME refactor and remove the second bool
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

func (db *DB) SearchQuotes(userID int64, query string, limit int32) ([]QuoteSearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()

	return db.q.SearchQuotes(ctx, base.SearchQuotesParams{UserID: userID, ToTsquery: query, Limit: limit})
}

func (db *DB) DeleteOutput(userID int64, outputChatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	err := db.q.DeleteOutput(ctx, base.DeleteOutputParams{UserID: userID, ChatID: outputChatID})
	return err
}

func (db *DB) DoCallback(user *User, callbackData *models.CallbackData) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	switch callbackData.Action {
	case models.CALLBACK_COMMAND_ACTIVATE_OUTPUT:
		outputChatID, err := strconv.ParseInt(callbackData.Data, 10, 0)
		if err != nil {
			return err
		}
		_, err = db.q.ActivateOutput(ctx, base.ActivateOutputParams{ChatID: outputChatID, UserID: user.ID})
		if err != nil {
			return err
		}
	case models.CALLBACK_COMMAND_DEACTIVATE_OUTPUT:
		outputChatID, err := strconv.ParseInt(callbackData.Data, 10, 0)
		if err != nil {
			return err
		}
		_, err = db.q.DeactivateOutput(ctx, base.DeactivateOutputParams{ChatID: outputChatID, UserID: user.ID})
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown callback data action")
	}

	return nil
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

	return nil
}

func (db *DB) Close() {
	db.pool.Close()
}

package bot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	u "github.com/aigic8/warmlight/bot/utils"
	"github.com/aigic8/warmlight/db"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TODO use as config to newBot
const DEFAULT_ACTIVE_SOURCE_TIMEOUT = 60
const DEACTIVATOR_INTERVAL_MINS = 10

func NewBot(DB *db.DB, token string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	h := Handlers{DB: DB}
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(h.updateHandler),
	}

	b, err := bot.New(os.Getenv(token), opts...)
	if err != nil {
		return err
	}

	deactivator := NewSourceDeactiver(DB, b, ctx)
	deactivator.Schedule(DEACTIVATOR_INTERVAL_MINS)

	b.Start(ctx)

	return nil
}

type Handlers struct {
	DB *db.DB
}

func (h Handlers) updateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil || update.Message.From.IsBot {
		return
	}

	user, userCreated, err := h.DB.GetOrCreateUser(uint(update.Message.From.ID), uint(update.Message.Chat.ID), update.Message.From.FirstName)
	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		logErr(err)
		return
	}

	var r u.Reaction
	switch {
	case userCreated:
		r, err = h.reactNewUser(user, update)
	case update.Message.Text == "/start":
		r, err = h.reactAlreadyJoinedStart(user, update)
	case strings.HasPrefix(update.Message.Text, "/setActiveSource"):
		r, err = h.reactSetActiveSource(user, update)
	default:
		r, err = h.reactStateNormal(user, update)
	}

	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		logErr(err)
		return
	}

	if err = r.Do(ctx, b); err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		logErr(err)
		return
	}
}

/////////////////////// REACTIONS ////////////////////////////

func (h Handlers) reactStateNormal(user *db.User, update *models.Update) (u.Reaction, error) {
	q, err := u.ParseQuote(update.Message.Text)
	if err != nil {
		return u.Reaction{}, err
	}

	_, err = h.DB.CreateQuoteWithData(uint(update.Message.From.ID), q.Text, q.MainSource, q.Tags, q.Sources)
	if err != nil {
		return u.Reaction{}, err
	}

	return u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, strQuoteAdded),
		},
	}, nil
}

func (h Handlers) reactNewUser(user *db.User, update *models.Update) (u.Reaction, error) {
	var messageText string
	if update.Message.Text == "/start" {
		messageText = strWelcomeToBot(user.FirstName)
	} else {
		messageText = strYourDataIsLost(user.FirstName)
	}

	return u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, messageText),
		},
	}, nil
}

func (h Handlers) reactAlreadyJoinedStart(user *db.User, update *models.Update) (u.Reaction, error) {
	return u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, strYouAreAlreadyJoined(user.FirstName)),
		},
	}, nil
}

func (h Handlers) reactSetActiveSource(user *db.User, update *models.Update) (u.Reaction, error) {
	// TODO what heppens if we already have an active source?
	text := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/setActiveSource"))
	argsRaw := strings.Split(text, ",")
	args := make([]string, 0, len(argsRaw))

	for _, arg := range argsRaw {
		args = append(args, strings.TrimSpace(arg))
	}

	argsLen := len(args)

	malformedReaction := u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, strMalformedSetActiveSource),
		},
	}

	if text == "" || argsLen > 2 {
		return malformedReaction, nil
	}

	activeSourceTimeoutInt := DEFAULT_ACTIVE_SOURCE_TIMEOUT
	var err error
	if argsLen == 2 {
		activeSourceTimeoutInt, err = strconv.Atoi(args[1])
		if err != nil {
			return malformedReaction, nil
		}
	}

	if activeSourceTimeoutInt <= 0 {
		return u.Reaction{
			Messages: []bot.SendMessageParams{
				u.TextReplyToMessage(update.Message, strSourceTimeoutShouldBeGreaterThanZero),
			},
		}, nil
	}
	activeSourceExpire := time.Now().Add(time.Minute * time.Duration(activeSourceTimeoutInt))

	_, err = h.DB.GetSource(user.ID, args[0])
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return u.Reaction{
				Messages: []bot.SendMessageParams{
					u.TextReplyToMessage(update.Message, strSourceDoesExist(args[0])),
				},
			}, nil
		}

		return u.Reaction{}, err
	}

	effected, err := h.DB.SetActiveSource(user.ID, args[0], activeSourceExpire)
	if err != nil {
		return u.Reaction{}, err
	}
	if !effected {
		return u.Reaction{}, errors.New("setActiveSource does not have any errors does not effect any rows")
	}

	return u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, strActiveSourceIsSet(args[0], activeSourceTimeoutInt)),
		},
	}, nil
}

// TODO better logging errors!
func logErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

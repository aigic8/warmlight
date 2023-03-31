package bot

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	u "github.com/aigic8/warmlight/bot/utils"
	"github.com/aigic8/warmlight/db"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func NewBot(DB *db.DB, token string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	h := Handlers{DB: DB}
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(h.updateHandler),
	}

	b, _ := bot.New(os.Getenv(token), opts...)
	b.Start(ctx)
}

type Handlers struct {
	DB *db.DB
}

func (h Handlers) updateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil || update.Message.From.IsBot {
		return
	}

	user, userCreated, err := h.DB.GetOrCreateUser(uint(update.Message.From.ID), update.Message.From.FirstName)
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
	case user.State == "normal":
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

// TODO better logging errors!
func logErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

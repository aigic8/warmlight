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
	"github.com/rs/zerolog"
)

// TODO use as config to newBot
const DEFAULT_ACTIVE_SOURCE_TIMEOUT = 60
const DEACTIVATOR_INTERVAL_MINS = 10
const IS_DEV = true
const LOG_PATH = "log/warmlight.log"

// TODO add support for filtering HASHTAGS and SOURCES for different outputs

func NewBot(appDB *db.DB, token string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	l, err := u.NewBotLogger(IS_DEV, LOG_PATH)
	if err != nil {
		return err
	}

	h := Handlers{db: appDB, l: l}
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(h.updateHandler),
	}

	b, err := bot.New(os.Getenv(token), opts...)
	if err != nil {
		return err
	}

	deactivator, err := NewSourceDeactiver(appDB, b, ctx)
	if err != nil {
		return err
	}
	deactivator.Schedule(DEACTIVATOR_INTERVAL_MINS)

	b.Start(ctx)

	return nil
}

type Handlers struct {
	db *db.DB
	l  zerolog.Logger
}

func (h Handlers) updateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO support groups
	if update.MyChatMember != nil {
		if _, err := h.reactMyChatMember(update); err != nil {
			h.l.Error().Err(err).Msg("reacting to chat member update")
		}
		return
	}

	if update.InlineQuery != nil {
		results, err := h.reactInlineQuery(update)
		if err != nil {
			h.l.Error().Err(err).Msg("reacting to inline query")
			return
		}

		if len(results) == 0 {
			return
		}

		success, err := b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			Results:       results,
		})
		if err != nil {
			h.l.Error().Err(err).Msg("answering inline query")
			return
		}
		if !success {
			h.l.Error().Msg("answering inline query false success")
			return
		}

		return
	}

	if update.Message == nil || update.Message.From == nil || update.Message.From.IsBot || update.Message.Chat.Type != "private" {
		return
	}

	user, userCreated, err := h.db.GetOrCreateUser(update.Message.From.ID, update.Message.Chat.ID, update.Message.From.FirstName)
	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		h.l.Error().Err(err).Msg("getting or creating user")
		return
	}

	var r u.Reaction
	switch {
	case userCreated:
		r, err = h.reactNewUser(user, update)
	case update.Message.Text == COMMAND_START:
		r, err = h.reactAlreadyJoinedStart(user, update)
	case strings.HasPrefix(update.Message.Text, COMMAND_SET_ACTIVE_SOURCE):
		r, err = h.reactSetActiveSource(user, update)
	case strings.HasPrefix(update.Message.Text, COMMAND_ADD_OUTPUT):
		r, err = h.reactAddOutput(user, update)
	default:
		r, err = h.reactDefault(user, update)
	}

	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		h.l.Error().Err(err).Msg("sending internal server error message")
		return
	}

	if err = r.Do(ctx, b); err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   strInternalServerErr,
		})
		h.l.Error().Err(err).Msg("sending reaction messages")
		return
	}
}

/////////////////////// REACTIONS ////////////////////////////

// TODO split reactions to multiple files
func (h Handlers) reactDefault(user *db.User, update *models.Update) (u.Reaction, error) {
	// FIXME test reactDefault with outputs
	q, err := u.ParseQuote(update.Message.Text)
	if err != nil {
		return u.Reaction{}, err
	}

	messages := []bot.SendMessageParams{}

	if user.ActiveSource.Valid {
		if len(q.Sources) == 0 {
			q.MainSource = user.ActiveSource.String
			q.Sources = append(q.Sources, user.ActiveSource.String)
		}
		if user.ActiveSourceExpire.Time.Before(time.Now()) {
			h.db.DeactivateSource(user.ID)
			messages = append(messages, bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   strActiveSourceExpired,
			})
		}
	}

	_, err = h.db.CreateQuoteWithData(int64(update.Message.From.ID), q.Text, q.MainSource, q.Tags, q.Sources)
	if err != nil {
		return u.Reaction{}, err
	}

	outputs, err := h.db.GetOutputs(int64(update.Message.From.ID))
	if err != nil {
		return u.Reaction{
			Messages: []bot.SendMessageParams{
				u.TextReplyToMessage(update.Message, strQuoteAddedButFailedToPublish),
			},
		}, nil
	}

	messages = append(messages, u.TextReplyToMessage(update.Message, strQuoteAdded))
	for _, output := range outputs {
		messages = append(messages, bot.SendMessageParams{
			ChatID:    output.ChatID,
			ParseMode: models.ParseModeMarkdown,
			Text:      strQuote(q),
		})
	}

	return u.Reaction{
		Messages: messages,
	}, nil
}

func (h Handlers) reactNewUser(user *db.User, update *models.Update) (u.Reaction, error) {
	var messageText string
	if update.Message.Text == COMMAND_START {
		messageText = strWelcomeToBot(user.Firstname)
	} else {
		messageText = strYourDataIsLost(user.Firstname)
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
			u.TextReplyToMessage(update.Message, strYouAreAlreadyJoined(user.Firstname)),
		},
	}, nil
}

func (h Handlers) reactSetActiveSource(user *db.User, update *models.Update) (u.Reaction, error) {
	// TODO what heppens if we already have an active source?
	text := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, COMMAND_SET_ACTIVE_SOURCE))
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

	_, err = h.db.GetSource(user.ID, args[0])
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

	_, effected, err := h.db.SetActiveSource(user.ID, args[0], activeSourceExpire)
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

func (h Handlers) reactAddOutput(user *db.User, update *models.Update) (u.Reaction, error) {
	// TODO philosophy question: what to use to identify the output?
	// some outputs might not have username and there is the possibility of two outputs having the same title
	// How to fix this?
	// 1. ask user to forward a message from the channel, that way we can have the chat ID
	// 2. List all the channels from user with buttons and ask user to click the channel button
	// FIXME handle output title change!
	chatTitle := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, COMMAND_ADD_OUTPUT))
	output, err := h.db.GetOutput(user.ID, chatTitle)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return u.Reaction{
				Messages: []bot.SendMessageParams{
					u.TextReplyToMessage(update.Message, strOutputNotFound(chatTitle)),
				},
			}, nil
		}
		return u.Reaction{}, err
	}

	if output.IsActive {
		return u.Reaction{
			Messages: []bot.SendMessageParams{
				u.TextReplyToMessage(update.Message, strOutputIsAlreadyActive(chatTitle)),
			},
		}, nil
	}

	_, err = h.db.SetOutputActive(user.ID, chatTitle)
	if err != nil {
		return u.Reaction{}, err
	}

	return u.Reaction{
		Messages: []bot.SendMessageParams{
			u.TextReplyToMessage(update.Message, strOutputIsSet(chatTitle)),
		},
	}, nil
}

func (h Handlers) reactMyChatMember(update *models.Update) (u.Reaction, error) {
	// FIXME test reactMyChatMemeber
	chat := update.MyChatMember.Chat
	from := update.MyChatMember.From
	// TODO support groups
	if chat.Type != "channel" {
		return u.Reaction{}, nil
	}

	adminChatMemeber := update.MyChatMember.NewChatMember.Administrator
	ownerChatMember := update.MyChatMember.NewChatMember.Owner
	if (adminChatMemeber != nil && adminChatMemeber.CanPostMessages) || ownerChatMember != nil {
		if _, _, err := h.db.GetOrCreateOutput(int64(from.ID), int64(chat.ID), chat.Title); err != nil {
			return u.Reaction{}, err
		}
	} else {
		if err := h.db.DeleteOutput(int64(from.ID), chat.Title); err != nil {
			return u.Reaction{}, err
		}
	}

	return u.Reaction{}, nil
}

func (h Handlers) reactInlineQuery(update *models.Update) ([]models.InlineQueryResult, error) {
	// TODO test reactInlineQuery
	if update.InlineQuery.From.IsBot {
		return nil, nil
	}

	user, err := h.db.GetUser(update.InlineQuery.From.ID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}
		return nil, nil
	}

	query := strings.Join(strings.Fields(update.InlineQuery.Query), " & ")
	// https://core.telegram.org/bots/api#answerinlinequery no more than 50 results per query is allowed
	quotes, err := h.db.SearchQuotes(user.ID, query, 50)
	if err != nil {
		return nil, err
	}

	// TODO add tags
	results := make([]models.InlineQueryResult, len(quotes))
	for _, q := range quotes {
		title := "No Source"
		description := q.Text
		if len(q.Text) > 40 {
			description = q.Text[:37] + "..."
		}
		if q.MainSource.Valid {
			title = q.MainSource.String
		}
		results = append(results, &models.InlineQueryResultArticle{
			ID:          fmt.Sprintf("%d", q.ID),
			Title:       title,
			Description: description,
			InputMessageContent: models.InputTextMessageContent{
				MessageText: strQuote(&u.Quote{Text: q.Text, MainSource: q.MainSource.String}),
				ParseMode:   models.ParseModeMarkdown,
			},
		})
	}

	return results, nil
}

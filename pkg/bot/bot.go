package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	m "github.com/aigic8/warmlight/pkg/bot/models"
	s "github.com/aigic8/warmlight/pkg/bot/strs"
	u "github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/hako/durafmt"
	"github.com/jackc/pgtype"
	"github.com/rs/zerolog"
)

type Config struct {
	WebhookAddress                 string
	CertFilePath                   string
	PrivKeyFilePath                string
	IsDev                          bool
	LogPath                        string
	DefaultActiveSourceTimeoutMins int
	DeactivatorIntervalMins        int
	Port                           int
}

const SOURCES_PAGE_LIMIT = 5

// TODO: add support for filtering HASHTAGS and SOURCES for different outputs

func RunBot(appDB *db.DB, token string, config *Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	l, err := u.NewBotLogger(config.IsDev, config.LogPath)
	if err != nil {
		return err
	}

	h := Handlers{
		db:                             appDB,
		l:                              l,
		defaultActiveSourceTimeoutMins: config.DefaultActiveSourceTimeoutMins,
		LibraryUUIDLifetime:            time.Duration(config.DefaultActiveSourceTimeoutMins) * time.Minute,
	}
	opts := []bot.Option{
		bot.WithDefaultHandler(h.updateHandler),
	}

	if config.IsDev {
		opts = append(opts, bot.WithDebug())
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return err
	}

	deactivator, err := NewSourceDeactiver(appDB, b, config.IsDev, config.LogPath, ctx)
	if err != nil {
		return err
	}
	deactivator.Schedule(config.DeactivatorIntervalMins)

	go b.StartWebhook(ctx)

	_, err = b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: config.WebhookAddress,
	})

	if err != nil {
		panic(err)
	}

	if err = http.ListenAndServeTLS(fmt.Sprintf(":%d", config.Port), config.CertFilePath, config.PrivKeyFilePath, b.WebhookHandler()); err != nil {
		panic(err)
	}

	return nil
}

type Handlers struct {
	db                             *db.DB
	l                              zerolog.Logger
	defaultActiveSourceTimeoutMins int
	LibraryUUIDLifetime            time.Duration
}

func (h Handlers) updateHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: support groups
	if update.MyChatMember != nil {
		if _, err := h.reactMyChatMember(update); err != nil {
			h.l.Error().Err(err).Msg("reacting to chat member update")
		}
		return
	}

	var r u.Reaction
	if update.CallbackQuery != nil {
		r, err := h.reactCallbackQuery(update)
		if err != nil {
			h.l.Error().Err(err).Msg("reacting to callback query")
			return
		}
		if err = r.Do(ctx, b); err != nil {
			h.l.Error().Err(err).Msg("sending reaction messages to callback query")
			_, err = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   s.InternalServerErr,
			})
			if err != nil {
				h.l.Error().Err(err).Msg("sending internal server error message")
			}
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
			Text:   s.InternalServerErr,
		})
		h.l.Error().Err(err).Msg("getting or creating user")
		return
	}

	switch {
	case userCreated:
		r, err = h.reactNewUser(user, update)
	case user.State == db.UserStateEditingSource:
		r, err = h.reactStateEditingSource(user, update)
	case user.State == db.UserStateConfirmingLibraryChange:
		r, err = h.reactStateConfirmingLibraryChange(user, update)
	case update.Message.Text == s.COMMAND_START:
		r, err = h.reactAlreadyJoinedStart(user, update)
	case update.Message.Text == s.COMMAND_HELP:
		r, err = h.reactHelp(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_SET_ACTIVE_SOURCE):
		r, err = h.reactSetActiveSource(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_DEACTIVATE_SOURCE):
		r, err = h.reactDeactivateSource(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_GET_OUTPUTS):
		r, err = h.reactGetOutputs(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_GET_LIBRARY_TOKEN):
		r, err = h.reactGetLibraryToken(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_SET_LIBRARY_TOKEN):
		r, err = h.reactSetLibraryToken(user, update)
	case strings.HasPrefix(update.Message.Text, s.COMMAND_GET_SOURCES):
		r, err = h.reactGetSources(user, update)
	default:
		r, err = h.reactDefault(user, update)
	}

	if err != nil {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   s.InternalServerErr,
		})
		h.l.Error().Err(err).Msg("sending internal server error message")
		return
	}

	if err = r.Do(ctx, b); err != nil {
		h.l.Error().Err(err).Msg("sending messages")
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   s.InternalServerErr,
		})
		h.l.Error().Err(err).Msg("sending internal server error message")
		return
	}
}

/////////////////////// REACTIONS ////////////////////////////

// TODO: split reactions to multiple files
func (h Handlers) reactDefault(user *db.User, update *models.Update) (u.Reaction, error) {
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
				Text:   s.ActiveSourceExpired,
			})
		}
	}

	_, err = h.db.CreateQuoteWithData(user.LibraryID, q.Text, q.MainSource, q.Tags, q.Sources)
	if err != nil {
		return u.Reaction{}, err
	}

	outputs, err := h.db.GetOutputs(update.Message.From.ID)
	if err != nil {
		return u.ReplyReaction(update.Message, s.QuoteAddedButFailedToPublish), nil
	}

	messages = append(messages, u.TextReplyToMessage(update.Message, s.QuoteAdded))
	for _, output := range outputs {
		messages = append(messages, bot.SendMessageParams{
			ChatID:    output.ChatID,
			ParseMode: models.ParseModeMarkdown,
			Text:      s.Quote(q),
		})
	}

	return u.Reaction{
		Messages: messages,
	}, nil
}

func (h Handlers) reactNewUser(user *db.User, update *models.Update) (u.Reaction, error) {
	var messageText string
	if update.Message.Text == s.COMMAND_START {
		messageText = s.WelcomeToBot(user.FirstName)
	} else {
		messageText = s.YourDataIsLost
	}

	return u.ReplyReaction(update.Message, messageText), nil
}

func (h Handlers) reactStateEditingSource(user *db.User, update *models.Update) (u.Reaction, error) {
	// TODO: use strings
	if update.Message.Text == "cancel" {
		if _, err := h.db.SetUserStateNormal(user.ID); err != nil {
			return u.Reaction{}, err
		}
		return u.Reaction{Messages: []bot.SendMessageParams{u.TextReplyToMessage(update.Message, s.OperationCanceled)}}, nil
	}

	var stateData db.StateEditingSourceData
	if err := json.Unmarshal(user.StateData.Bytes, &stateData); err != nil {
		if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
			return u.Reaction{}, err
		}
		return u.TextReaction(update.Message.Chat.ID, s.GoingBackToNormalMode), nil
	}

	currentSource, err := h.db.GetSourceByID(user.LibraryID, stateData.SourceID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
				return u.Reaction{}, err
			}
			return u.Reaction{Messages: []bot.SendMessageParams{
				u.TextReplyToMessage(update.Message, s.SourceNoLongerExists),
				u.TextMessage(update.Message.Chat.ID, s.GoingBackToNormalMode),
			}}, nil
		}
		return u.Reaction{}, err
	}

	editMap := map[string]string{}
	for _, line := range strings.Split(update.Message.Text, "\n") {
		if line == "" {
			continue
		}
		lineParts := strings.SplitN(line, ":", 2)
		if len(lineParts) < 2 {
			return u.Reaction{Messages: []bot.SendMessageParams{u.TextMessage(update.Message.Chat.ID, s.MalformedEditSourceText)}}, nil
		}
		editMap[strings.TrimSpace(lineParts[0])] = strings.TrimSpace(lineParts[1])
	}

	kind := currentSource.Kind
	if sourceKind, ok := editMap[s.SOURCE_KIND]; ok {
		if !u.IsValidSourceKind(sourceKind) {
			return u.ReplyReaction(update.Message, s.InvalidSourceKind(editMap["kind"])), nil
		}
		kind = db.SourceKind(sourceKind)
	}

	newSource := *currentSource
	switch kind {
	case db.SourceKindUnknown:
		newSource.Kind = db.SourceKindUnknown
		newSource.Data = pgtype.JSON{Status: pgtype.Null}
	case db.SourceKindBook:
		newSource.Kind = db.SourceKindBook
		jsonBytes, err := u.EditMapToJsonBook(currentSource, editMap)
		if err != nil {
			return u.Reaction{}, err
		}
		newSource.Data = pgtype.JSON{Bytes: jsonBytes, Status: pgtype.Present}
	case db.SourceKindPerson:
		newSource.Kind = db.SourceKindPerson
		jsonBytes, err := u.EditMapToJsonPerson(currentSource, editMap)
		if err != nil {
			if errors.Is(err, u.ErrMalformedDates) {
				return u.ReplyReaction(update.Message, s.MalformedPersonDates), nil
			}
			return u.Reaction{}, err
		}
		newSource.Data = pgtype.JSON{Bytes: jsonBytes, Status: pgtype.Present}
	case db.SourceKindArticle:
		newSource.Kind = db.SourceKindArticle
		jsonBytes, err := u.EditMapToJsonArticle(currentSource, editMap)
		if err != nil {
			return u.Reaction{}, err
		}
		newSource.Data = pgtype.JSON{Bytes: jsonBytes, Status: pgtype.Present}
	}

	if sourceName, ok := editMap[s.SOURCE_NAME]; ok {
		newSource.Name = sourceName
	}

	resSource, err := h.db.UpdateSource(user.LibraryID, &newSource)
	if err != nil {
		return u.Reaction{}, nil
	}

	updateStr, err := s.UpdatedSource(resSource)
	if err != nil {
		return u.Reaction{}, nil
	}

	if _, err = h.db.SetUserStateNormal(update.Message.From.ID); err != nil {
		return u.Reaction{}, err
	}

	return u.ReplyReaction(update.Message, updateStr), nil
}

func (h Handlers) reactStateConfirmingLibraryChange(user *db.User, update *models.Update) (u.Reaction, error) {
	if update.Message.Text == s.ConfirmLibraryChangeYesAnswer {
		var stateData db.StateConfirmingLibraryChangeData
		if err := json.Unmarshal(user.StateData.Bytes, &stateData); err != nil {
			return u.Reaction{}, err
		}

		_, err := h.db.GetLibrary(stateData.LibraryID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
					return u.Reaction{}, err
				}
				return u.TextReaction(update.Message.Chat.ID, s.LibraryNoLongerExistsOPCanceled), nil
			}

			return u.Reaction{}, err
		}

		if stateData.Mode == db.ChangeLibraryDeleteMode {
			err = h.db.DeleteUserCurrentLibraryAndMigrateTo(user.ID, user.LibraryID, stateData.LibraryID)
			if err != nil {
				return u.Reaction{}, err
			}
		} else if stateData.Mode == db.ChangeLibraryMergeMode {
			err = h.db.MergeUserCurrentLibraryAndMigrateTo(user.ID, user.LibraryID, stateData.LibraryID)
			if err != nil {
				return u.Reaction{}, err
			}
		} else {
			return u.Reaction{}, fmt.Errorf("unknown change library mode: '%s' ", stateData.Mode)
		}

		if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
			return u.Reaction{}, fmt.Errorf("setting user state normal: %w", err)
		}

		return u.ReplyReaction(update.Message, s.LibraryChangedSuccessfully), nil
	}

	if update.Message.Text == s.ConfirmLibraryChangeCancelAnswer {
		if _, err := h.db.SetUserStateNormal(user.ID); err != nil {
			return u.Reaction{}, err
		}
		return u.ReplyReaction(update.Message, s.OperationCanceled), nil
	}

	return u.ReplyReaction(update.Message, s.UnknownLibraryConfirmationMessage), nil
}

func (h Handlers) reactAlreadyJoinedStart(user *db.User, update *models.Update) (u.Reaction, error) {
	return u.ReplyReaction(update.Message, s.YouAreAlreadyJoined), nil
}

func (h Handlers) reactGetSources(user *db.User, update *models.Update) (u.Reaction, error) {
	text := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, s.COMMAND_GET_SOURCES))

	filter, err := m.ParseSourceFilter(text)
	if err != nil {
		if errors.Is(err, m.ErrMultipleSourceKindFilters) {
			return u.ReplyReaction(update.Message, s.OnlyOneSourceKindFilterIsAllowed), nil
		}
		return u.Reaction{}, err
	}

	// we are fetching one more elements to check if there is any more pages or not
	sources, err := h.db.QuerySources(db.QuerySourcesParams{
		LibraryID:  user.LibraryID,
		NameQuery:  filter.Text,
		SourceKind: filter.SourceKind,
		Limit:      SOURCES_PAGE_LIMIT + 1,
		BaseID:     0,
		Before:     false,
	})
	if err != nil {
		return u.Reaction{}, err
	}

	sourcesLen := len(sources)
	isLastPage := sourcesLen < SOURCES_PAGE_LIMIT+1
	var viewedSources []db.Source
	if !isLastPage {
		// excluding the last element in viewed sources
		viewedSources = sources[:sourcesLen-1]
	} else {
		viewedSources = sources[:]
	}

	msg := u.TextReplyToMessage(update.Message, s.ListOfSources(viewedSources))
	msg.ParseMode = models.ParseModeMarkdown
	msg.ReplyMarkup = u.SourcesReplyMarkup(viewedSources, true, isLastPage)
	return u.Reaction{Messages: []bot.SendMessageParams{msg}}, nil
}

func (h Handlers) reactSetActiveSource(user *db.User, update *models.Update) (u.Reaction, error) {
	// TODO: what happens if we already have an active source?
	text := strings.TrimSpace(strings.TrimPrefix(update.Message.Text, s.COMMAND_SET_ACTIVE_SOURCE))
	argsRaw := strings.Split(text, ",")
	args := make([]string, 0, len(argsRaw))

	for _, arg := range argsRaw {
		args = append(args, strings.TrimSpace(arg))
	}

	argsLen := len(args)

	malformedReaction := u.ReplyReaction(update.Message, s.MalformedSetActiveSource(h.defaultActiveSourceTimeoutMins))
	if text == "" || argsLen > 2 {
		return malformedReaction, nil
	}

	activeSourceTimeoutInt := h.defaultActiveSourceTimeoutMins
	var err error
	if argsLen == 2 {
		activeSourceTimeoutInt, err = strconv.Atoi(args[1])
		if err != nil {
			return malformedReaction, nil
		}
	}

	if activeSourceTimeoutInt <= 0 {
		return u.ReplyReaction(update.Message, s.SourceTimeoutShouldBeGreaterThanZero), nil
	}
	activeSourceExpire := time.Now().Add(time.Minute * time.Duration(activeSourceTimeoutInt))

	_, err = h.db.GetSource(user.LibraryID, args[0])
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return u.ReplyReaction(update.Message, s.SourceDoesNotExist(args[0])), nil
		}

		return u.Reaction{}, err
	}

	_, err = h.db.SetActiveSource(user.ID, args[0], activeSourceExpire)
	if err != nil {
		return u.Reaction{}, err
	}

	return u.ReplyReaction(update.Message, s.ActiveSourceIsSet(args[0], activeSourceTimeoutInt)), nil
}

func (h Handlers) reactHelp(user *db.User, update *models.Update) (u.Reaction, error) {
	return u.TextReaction(update.Message.Chat.ID, s.Help), nil
}

func (h Handlers) reactGetOutputs(user *db.User, update *models.Update) (u.Reaction, error) {
	// FIXME: paginate get outputs
	outputs, err := h.db.GetOutputs(user.ID)
	if err != nil {
		return u.Reaction{}, err
	}

	msg := u.TextReplyToMessage(update.Message, s.ListOfYourOutputs(outputs))
	msg.ParseMode = models.ParseModeMarkdown

	replyMarkup, err := u.OutputsReplyMarkup(outputs)
	if err != nil {
		return u.Reaction{}, err
	}

	msg.ReplyMarkup = replyMarkup
	return u.Reaction{Messages: []bot.SendMessageParams{msg}}, nil
}

func (h Handlers) reactGetLibraryToken(user *db.User, update *models.Update) (u.Reaction, error) {
	library, err := h.db.GetLibrary(user.LibraryID)
	if err != nil {
		return u.Reaction{}, err
	}

	if library.OwnerID != user.ID {
		return u.ReplyReaction(update.Message, s.OnlyTheOwnerCanAddNewUsers), nil
	}

	UUID := uuid.New()
	_, err = h.db.SetLibraryToken(user.LibraryID, UUID, time.Now().Add(h.LibraryUUIDLifetime))
	if err != nil {
		return u.Reaction{}, err
	}

	UUIDLifetimeStr := durafmt.Parse(h.LibraryUUIDLifetime).String()
	return u.ReplyReaction(update.Message, s.YourLibraryToken(UUID.String(), UUIDLifetimeStr)), nil
}

func (h Handlers) reactSetLibraryToken(user *db.User, update *models.Update) (u.Reaction, error) {
	textParts := strings.Fields(update.Message.Text)
	if len(textParts) != 2 {
		return u.ReplyReaction(update.Message, s.MalformedLibraryToken), nil
	}

	libraryUUID, err := uuid.Parse(textParts[1])
	if err != nil {
		return u.ReplyReaction(update.Message, s.MalformedLibraryToken), nil
	}

	library, err := h.db.GetLibraryByUUID(libraryUUID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return u.ReplyReaction(update.Message, s.NoLibraryExistsWithToken), nil
		}

		return u.Reaction{}, err
	}

	if library.TokenExpiresOn.Time.Before(time.Now()) {
		return u.ReplyReaction(update.Message, s.LibraryTokenExpired), nil
	}

	if _, err = h.db.SetUserStateChangingLibrary(user.ID, library.ID); err != nil {
		return u.Reaction{}, err
	}

	message := u.TextReplyToMessage(update.Message, s.MergeOrDeleteCurrentLibraryData)
	message.ReplyMarkup = u.MergeOrDeleteCurrentLibraryReplyMarkup
	return u.Reaction{
		Messages: []bot.SendMessageParams{message},
	}, nil
}

func (h Handlers) reactMyChatMember(update *models.Update) (u.Reaction, error) {
	// TODO: test reactMyChatMember
	chat := update.MyChatMember.Chat
	from := update.MyChatMember.From
	// TODO: support groups
	if chat.Type != "channel" {
		return u.Reaction{}, nil
	}

	adminChatMember := update.MyChatMember.NewChatMember.Administrator
	ownerChatMember := update.MyChatMember.NewChatMember.Owner
	if (adminChatMember != nil && adminChatMember.CanPostMessages) || ownerChatMember != nil {
		if _, _, err := h.db.GetOrCreateOutput(from.ID, chat.ID, chat.Title); err != nil {
			return u.Reaction{}, err
		}
	} else {
		if err := h.db.DeleteOutput(from.ID, chat.ID); err != nil {
			return u.Reaction{}, err
		}
	}

	return u.Reaction{}, nil
}

func (h Handlers) reactDeactivateSource(user *db.User, update *models.Update) (u.Reaction, error) {
	activeSource := user.ActiveSource.String
	if !user.ActiveSource.Valid {
		return u.ReplyReaction(update.Message, s.NoActiveSource), nil
	}
	_, err := h.db.DeactivateSource(user.ID)
	if err != nil {
		return u.Reaction{}, err
	}

	return u.ReplyReaction(update.Message, s.ActiveSourceDeactivated(activeSource)), nil
}

func (h Handlers) reactCallbackQuery(update *models.Update) (u.Reaction, error) {
	// TODO: test reactCallbackQuery
	user, err := h.db.GetUser(update.CallbackQuery.Sender.ID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return u.Reaction{}, err
		}
		return u.Reaction{}, nil
	}

	callbackData, err := m.UnmarshalCallbackData(update.CallbackQuery.Data)
	if err != nil {
		return u.Reaction{}, err
	}

	if user.State == db.UserStateChangingLibrary {
		if callbackData.Action == m.CALLBACK_COMMAND_DELETE_LIBRARY || callbackData.Action == m.CALLBACK_COMMAND_MERGE_LIBRARY {
			var stateData db.StateChangingLibraryData
			if err = json.Unmarshal(user.StateData.Bytes, &stateData); err != nil {
				return u.Reaction{}, err
			}

			library, err := h.db.GetLibrary(stateData.LibraryID)
			if err != nil {
				if errors.Is(err, db.ErrNotFound) {
					if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
						return u.Reaction{}, err
					}
					return u.TextReaction(user.ChatID, s.LibraryNoLongerExistsOPCanceled), nil
				}
				return u.Reaction{}, err
			}

			if library.TokenExpiresOn.Time.Before(time.Now()) {
				if _, err = h.db.SetUserStateNormal(user.ID); err != nil {
					return u.Reaction{}, err
				}
				return u.TextReaction(user.ChatID, s.LibraryTokenExpired), nil
			}

			mode := db.ChangeLibraryMergeMode
			if callbackData.Action == m.CALLBACK_COMMAND_DELETE_LIBRARY {
				mode = db.ChangeLibraryDeleteMode
			}

			_, err = h.db.SetUserStateConfirmingLibraryChange(user.ID, stateData.LibraryID, mode)
			if err != nil {
				return u.Reaction{}, err
			}

			text := s.ConfirmLibraryChange(s.ConfirmLibraryChangeYesAnswer, s.ConfirmLibraryChangeCancelAnswer)
			return u.TextReaction(user.ChatID, text), nil
		}
	}

	if callbackData.Action != "" {
		switch callbackData.Action {
		case m.CALLBACK_COMMAND_ACTIVATE_OUTPUT:
			outputChatID, err := strconv.ParseInt(callbackData.Data, 10, 0)
			if err != nil {
				return u.Reaction{}, err
			}
			if _, err = h.db.ActivateOutput(user.ID, outputChatID); err != nil {
				return u.Reaction{}, err
			}
		case m.CALLBACK_COMMAND_DEACTIVATE_OUTPUT:
			outputChatID, err := strconv.ParseInt(callbackData.Data, 10, 0)
			if err != nil {
				return u.Reaction{}, err
			}
			if _, err = h.db.DeactivateOutput(user.ID, outputChatID); err != nil {
				return u.Reaction{}, err
			}
		case m.CALLBACK_COMMAND_SOURCE_INFO:
			sourceID, err := strconv.ParseInt(callbackData.Data, 10, 0)
			if err != nil {
				return u.Reaction{}, err
			}
			source, err := h.db.GetSourceByID(user.LibraryID, sourceID)
			if err != nil {
				if errors.Is(err, db.ErrNotFound) {
					return u.TextReaction(user.ChatID, s.SourceNoLongerExists), nil
				}
				return u.Reaction{}, err
			}

			sourceData, err := u.ParseSourceData(source.Kind, source.Data)
			if err != nil {
				return u.Reaction{}, err
			}

			infoStr, err := s.SourceInfo(source, sourceData)
			if err != nil {
				return u.Reaction{}, err
			}
			return u.TextReaction(user.ChatID, infoStr), nil
		case m.CALLBACK_COMMAND_SOURCE_EDIT:
			sourceID, err := strconv.ParseInt(callbackData.Data, 10, 0)
			if err != nil {
				return u.Reaction{}, err
			}
			source, err := h.db.GetSourceByID(user.LibraryID, sourceID)
			if err != nil {
				if errors.Is(err, db.ErrNotFound) {
					return u.TextReaction(user.ChatID, s.SourceNoLongerExists), nil
				}
				return u.Reaction{}, err
			}

			if _, err = h.db.SetUserStateEditingSource(user.ID, sourceID); err != nil {
				return u.Reaction{}, err
			}

			sourceData, err := u.ParseSourceData(source.Kind, source.Data)
			if err != nil {
				return u.Reaction{}, err
			}

			msgText, err := s.EditSource(source, sourceData)
			if err != nil {
				return u.Reaction{}, err
			}

			return u.TextReaction(user.ChatID, msgText), nil
		default:
			return u.Reaction{}, errors.New("unknown callback data action")
		}
	}

	if callbackData.ReplaceMessageWith == m.CALLBACK_MSG_OUTPUTS_LIST {
		outputs, err := h.db.GetOutputs(user.ID)
		if err != nil {
			return u.Reaction{}, err
		}

		replyMarkup, err := u.OutputsReplyMarkup(outputs)
		if err != nil {
			return u.Reaction{}, err
		}

		return u.Reaction{
			EditMessages: []bot.EditMessageTextParams{
				{
					ChatID:      update.CallbackQuery.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.ID,
					Text:        s.ListOfYourOutputs(outputs),
					ParseMode:   models.ParseModeMarkdown,
					ReplyMarkup: replyMarkup,
				},
			},
		}, nil
	}

	return u.Reaction{}, nil
}

func (h Handlers) reactInlineQuery(update *models.Update) ([]models.InlineQueryResult, error) {
	// TODO: test reactInlineQuery
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
	quotes, err := h.db.SearchQuotes(user.LibraryID, query, 50)
	if err != nil {
		return nil, err
	}

	// TODO: add tags
	results := make([]models.InlineQueryResult, 0, len(quotes))
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
				MessageText: s.Quote(&u.Quote{Text: q.Text, MainSource: q.MainSource.String}),
				ParseMode:   models.ParseModeMarkdown,
			},
		})
	}

	return results, nil
}

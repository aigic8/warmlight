package utils

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/internal/db/base"
	m "github.com/aigic8/warmlight/pkg/bot/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgtype"
)

var ErrUnknownSourceKind = errors.New("unknown source kind")
var ErrMalformedDates = errors.New("malformed lived in dates")

func TextMessage(chatID int64, text string) bot.SendMessageParams {
	return bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
}

func TextReplyToMessage(message *models.Message, text string) bot.SendMessageParams {
	return bot.SendMessageParams{
		ReplyToMessageID: message.ID,
		ChatID:           message.Chat.ID,
		Text:             text,
	}
}

// returns callback command in format replaceMsgWith-command-data
func MakeToggleOutputStateCallback(output *db.Output) (string, error) {
	chatIDStr := strconv.FormatInt(output.ChatID, 10)
	callbackData := m.CallbackData{
		ReplaceMessageWith: m.CALLBACK_MSG_OUTPUTS_LIST,
		Action:             m.CALLBACK_COMMAND_ACTIVATE_OUTPUT,
		Data:               chatIDStr,
	}

	if output.IsActive {
		callbackData.Action = m.CALLBACK_COMMAND_DEACTIVATE_OUTPUT
	}

	return callbackData.Marshal(), nil
}

var mergeLibraryCallbackData = (&m.CallbackData{Action: m.CALLBACK_COMMAND_MERGE_LIBRARY}).Marshal()
var deleteLibraryCallbackData = (&m.CallbackData{Action: m.CALLBACK_COMMAND_DELETE_LIBRARY}).Marshal()

var MergeOrDeleteCurrentLibraryReplyMarkup = models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "merge", CallbackData: mergeLibraryCallbackData},
			{Text: "delete", CallbackData: deleteLibraryCallbackData},
		},
	},
}

func OutputsReplyMarkup(outputs []db.Output) (models.InlineKeyboardMarkup, error) {
	inlineKeyboard := [][]models.InlineKeyboardButton{}
	for _, output := range outputs {
		outputState := "deactive"
		if output.IsActive {
			outputState = "active"
		}
		callbackData, err := MakeToggleOutputStateCallback(&output)
		if err != nil {
			return models.InlineKeyboardMarkup{}, err
		}
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{Text: output.Title + " - " + outputState, CallbackData: callbackData}})
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}, nil
}

func MakeSourceKeyboardCallbacks(sourceID int64) (string, string) {
	infoCallbackData := m.CallbackData{
		Action:             m.CALLBACK_COMMAND_SOURCE_INFO,
		Data:               strconv.FormatInt(sourceID, 10),
		ReplaceMessageWith: "",
	}

	editCallbackData := m.CallbackData{
		Action:             m.CALLBACK_COMMAND_SOURCE_EDIT,
		Data:               strconv.FormatInt(sourceID, 10),
		ReplaceMessageWith: "",
	}

	return infoCallbackData.Marshal(), editCallbackData.Marshal()
}

func MakeSourceKeyboardPagesCallbacks(firstSourceID, lastSourceID int64) (string, string) {
	prevPageCallbackData := m.CallbackData{
		Data:               strconv.FormatInt(firstSourceID, 10),
		ReplaceMessageWith: m.CALLBACK_MSG_PREV_SOURCE_PAGE,
	}

	nextPageCallbackData := m.CallbackData{
		Data:               strconv.FormatInt(lastSourceID, 10),
		ReplaceMessageWith: m.CALLBACK_MSG_NEXT_SOURCE_PAGE,
	}

	return prevPageCallbackData.Marshal(), nextPageCallbackData.Marshal()
}

func IsValidSourceKind(sourceKind string) bool {
	for _, validSourceKind := range db.VALID_SOURCE_KINDS {
		if sourceKind == validSourceKind {
			return true
		}
	}

	return false
}

func EditMapToJsonBook(baseSource *db.Source, editMap map[string]string) ([]byte, error) {
	sourceData := db.SourceBookData{}
	if baseSource.Kind == db.SourceKindBook {
		baseSourceData, err := ParseSourceData(db.SourceKindBook, baseSource.Data)
		if err != nil {
			return nil, err
		}
		sourceData = baseSourceData.(db.SourceBookData)
	}

	// TODO: use base package strings
	if linkToInfo, exist := editMap["info url"]; exist {
		sourceData.LinkToInfo = linkToInfo
	}
	if author, exist := editMap["author"]; exist {
		sourceData.Author = author
	}
	if linkToAuthor, exist := editMap["author url"]; exist {
		sourceData.LinkToAuthor = linkToAuthor
	}

	return json.Marshal(&sourceData)
}

func EditMapToJsonPerson(baseSource *db.Source, editMap map[string]string) ([]byte, error) {
	sourceData := db.SourcePersonData{}
	if baseSource.Kind == db.SourceKindPerson {
		baseSourceData, err := ParseSourceData(db.SourceKindPerson, baseSource.Data)
		if err != nil {
			return nil, err
		}

		sourceData = baseSourceData.(db.SourcePersonData)
	}

	if linkToInfo, exist := editMap["info url"]; exist {
		sourceData.LinkToInfo = linkToInfo
	}
	if title, exist := editMap["title"]; exist {
		sourceData.Title = title
	}
	if livedIn, exist := editMap["lived in"]; exist {
		dates := strings.Split(livedIn, "-")
		if len(dates) != 2 {
			return nil, ErrMalformedDates
		}
		bornOn, err := strconv.Atoi(dates[0])
		if err != nil {
			return nil, ErrMalformedDates
		}

		deathOn, err := strconv.Atoi(dates[1])
		if err != nil {
			return nil, ErrMalformedDates
		}

		sourceData.BornOn = time.Date(bornOn, 1, 1, 1, 1, 1, 1, time.UTC)
		sourceData.DeathOn = time.Date(deathOn, 1, 1, 1, 1, 1, 1, time.UTC)
	}

	return json.Marshal(sourceData)
}

func EditMapToJsonArticle(baseSource *db.Source, editMap map[string]string) ([]byte, error) {
	sourceData := db.SourceArticleData{}
	if baseSource.Kind == base.SourceKindArticle {
		baseSourceData, err := ParseSourceData(db.SourceKindPerson, baseSource.Data)
		if err != nil {
			return nil, err
		}

		sourceData = baseSourceData.(db.SourceArticleData)
	}

	if linkToInfo, exist := editMap["info url"]; exist {
		sourceData.URL = linkToInfo
	}
	if author, exist := editMap["author"]; exist {
		sourceData.Author = author
	}

	return json.Marshal(sourceData)
}

func ParseSourceData(sourceKind db.SourceKind, sourceData pgtype.JSON) (any, error) {
	if sourceData.Status != pgtype.Present {
		return nil, nil
	}

	switch sourceKind {
	case db.SourceKindUnknown:
		return nil, nil
	case db.SourceKindBook:
		var data db.SourceBookData
		if err := json.Unmarshal(sourceData.Bytes, &data); err != nil {
			return nil, err
		}
		return data, nil
	case db.SourceKindPerson:
		var data db.SourcePersonData
		if err := json.Unmarshal(sourceData.Bytes, &data); err != nil {
			return nil, err
		}
		return data, nil
	case db.SourceKindArticle:
		var data db.SourceArticleData
		if err := json.Unmarshal(sourceData.Bytes, &data); err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, ErrUnknownSourceKind
	}
}

func SourcesReplyMarkup(sources []db.Source, firstPage, lastPage bool) models.InlineKeyboardMarkup {
	inlineKeyboard := [][]models.InlineKeyboardButton{}

	if len(sources) == 0 {
		return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
	}

	for i, source := range sources {
		num := i + 1
		numStr := strconv.Itoa(num)
		infoCallbackData, editCallbackData := MakeSourceKeyboardCallbacks(source.ID)
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{Text: numStr + ". Info", CallbackData: infoCallbackData},
			{Text: numStr + ". Edit", CallbackData: editCallbackData},
		})
	}

	firstSourceID := sources[0].ID
	lastSourceID := sources[len(sources)-1].ID
	prevPageCallbackData, nextPageCallbackData := MakeSourceKeyboardPagesCallbacks(firstSourceID, lastSourceID)

	if !firstPage && !lastPage {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{Text: "⬅️", CallbackData: prevPageCallbackData},
			{Text: "➡️", CallbackData: nextPageCallbackData},
		})
	} else if !firstPage {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{Text: "⬅️", CallbackData: prevPageCallbackData}})
	} else if !lastPage {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{Text: "➡️", CallbackData: nextPageCallbackData}})
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}

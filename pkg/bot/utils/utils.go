package utils

import (
	"strconv"

	"github.com/aigic8/warmlight/internal/db"
	m "github.com/aigic8/warmlight/pkg/bot/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

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

	if firstPage {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{Text: "➡️", CallbackData: nextPageCallbackData}})
	} else if lastPage {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{{Text: "⬅️", CallbackData: prevPageCallbackData}})
	} else {
		inlineKeyboard = append(inlineKeyboard, []models.InlineKeyboardButton{
			{Text: "⬅️", CallbackData: prevPageCallbackData},
			{Text: "➡️", CallbackData: nextPageCallbackData},
		})
	}

	return models.InlineKeyboardMarkup{InlineKeyboard: inlineKeyboard}
}

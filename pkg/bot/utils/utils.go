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
		ReplaceMessageWith: m.CALLBACK_OUTPUTS_LIST_MSG,
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

package utils

import (
	"encoding/json"

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

func MakeToggleOutputStateCallback(output *db.Output) (string, error) {
	var data m.CallbackData
	if output.IsActive {
		data = m.CallbackData{DeactivateOutputs: []int64{output.ChatID}, ReplaceMessageWith: m.OUTPUTS_LIST_MSG}
	} else {
		data = m.CallbackData{ActivateOutputs: []int64{output.ChatID}, ReplaceMessageWith: m.OUTPUTS_LIST_MSG}
	}

	dataBytes, err := json.Marshal(&data)
	if err != nil {
		return "", err
	}

	return string(dataBytes), nil
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

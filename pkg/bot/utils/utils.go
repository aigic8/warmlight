package utils

import (
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

package utils

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Reaction struct {
	Messages     []bot.SendMessageParams
	EditMessages []bot.EditMessageTextParams
}

func (r Reaction) Do(ctx context.Context, bot *bot.Bot) error {
	if r.Messages != nil {
		for _, msg := range r.Messages {
			_, err := bot.SendMessage(ctx, &msg)
			if err != nil {
				return err
			}
		}
	}

	if r.EditMessages != nil {
		for _, editedMessage := range r.EditMessages {
			_, err := bot.EditMessageText(ctx, &editedMessage)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ReplyReaction(message *models.Message, text string) Reaction {
	return Reaction{Messages: []bot.SendMessageParams{
		TextReplyToMessage(message, text),
	}}
}

func TextReaction(chatID int64, text string) Reaction {
	return Reaction{Messages: []bot.SendMessageParams{
		TextMessage(chatID, text),
	}}
}

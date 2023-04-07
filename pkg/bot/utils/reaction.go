package utils

import (
	"context"

	"github.com/go-telegram/bot"
)

// FIXME go framework agnostic way!
type Reaction struct {
	Messages []bot.SendMessageParams
}

func (r Reaction) Do(ctx context.Context, bot *bot.Bot) error {
	for _, msg := range r.Messages {
		_, err := bot.SendMessage(ctx, &msg)
		if err != nil {
			return err
		}
	}

	return nil
}

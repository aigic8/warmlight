package bot

import (
	"context"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	u "github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-co-op/gocron"
	"github.com/go-telegram/bot"
	"github.com/rs/zerolog"
)

type SourceDeactiver struct {
	db        *db.DB
	b         *bot.Bot
	ctx       context.Context
	scheduler *gocron.Scheduler
	l         zerolog.Logger
}

func NewSourceDeactiver(db *db.DB, b *bot.Bot, isDev bool, logPath string, ctx context.Context) (*SourceDeactiver, error) {
	l, err := u.NewSourceDeactiverLogger(isDev, logPath)
	if err != nil {
		return nil, err
	}
	return &SourceDeactiver{db: db, b: b, ctx: ctx, l: l}, nil
}

func (sd *SourceDeactiver) Schedule(intervalMins int) {
	sd.scheduler = gocron.NewScheduler(time.UTC)
	sd.scheduler.Every(intervalMins).Minutes().Do(sd.deactivateExpiredSources)
	sd.scheduler.StartAsync()
}

func (sd *SourceDeactiver) deactivateExpiredSources() {
	users, err := sd.db.DeactivateExpiredSources()
	if err != nil {
		sd.l.Error().Err(err).Msg("deactivating expiresed sources")
		return
	}

	for _, user := range users {
		_, err := sd.b.SendMessage(sd.ctx, &bot.SendMessageParams{
			ChatID: user.ChatID,
			Text:   strActiveSourceExpired,
		})

		if err != nil {
			sd.l.Error().Err(err).Msg("sending expired notifications to users")
		}
	}
}

func (sd *SourceDeactiver) Stop() {
	if sd.scheduler != nil {
		sd.scheduler.Stop()
	}
}

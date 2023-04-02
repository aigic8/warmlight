package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/aigic8/warmlight/db"
	"github.com/go-co-op/gocron"
	"github.com/go-telegram/bot"
)

type SourceDeactiver struct {
	db        *db.DB
	b         *bot.Bot
	ctx       context.Context
	scheduler *gocron.Scheduler
}

func NewSourceDeactiver(db *db.DB, b *bot.Bot, ctx context.Context) *SourceDeactiver {
	return &SourceDeactiver{db: db, b: b, ctx: ctx}
}

func (sd *SourceDeactiver) Schedule(intervalMins int) {
	sd.scheduler = gocron.NewScheduler(time.UTC)
	sd.scheduler.Every(intervalMins).Minutes().Do(sd.deactivateExpiredSources)
	sd.scheduler.StartAsync()
}

func (sd *SourceDeactiver) deactivateExpiredSources() {
	users, err := sd.db.DeactivateExpiredSources()
	if err != nil {
		// FIXME error handling
		fmt.Println(err)
		return
	}

	for _, user := range users {
		_, err := sd.b.SendMessage(sd.ctx, &bot.SendMessageParams{
			ChatID: user.ChatID,
			Text:   strActiveSourceExpired,
		})

		// TODO error handling
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (sd *SourceDeactiver) Stop() {
	if sd.scheduler != nil {
		sd.scheduler.Stop()
	}
}

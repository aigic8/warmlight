package main

import (
	"time"

	"github.com/aigic8/warmlight/bot"
	"github.com/aigic8/warmlight/db"
	"github.com/aigic8/warmlight/utils"
)

func main() {
	config, err := utils.LoadConfig("warmlight.sample.toml")
	if err != nil {
		panic(err)
	}

	db, err := db.NewDB(config.Db.URL, time.Duration(config.Db.TimeoutMs)*time.Millisecond)
	if err != nil {
		panic(err)
	}

	botConfig := &bot.Config{
		IsDev:                          config.Bot.IsDev,
		LogPath:                        config.Bot.LogPath,
		DefaultActiveSourceTimeoutMins: config.Bot.DefaultActiveSourceTimeoutMins,
		DeactivatorIntervalMins:        config.Bot.DeactivatorIntervalMins,
	}

	if err := bot.RunBot(db, config.Bot.Token, botConfig); err != nil {
		panic(err)
	}
}

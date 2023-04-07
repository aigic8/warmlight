package main

import (
	"os"
	"path"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot"
	"github.com/aigic8/warmlight/pkg/utils"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config, err := utils.LoadConfig(path.Join(cwd, "warmlight.toml"))
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
		WebhookAddress:                 config.Bot.WebhookURL,
		CertFilePath:                   config.Bot.CertFilePath,
		PrivKeyFilePath:                config.Bot.PrivKeyFilePath,
		DefaultActiveSourceTimeoutMins: config.Bot.DefaultActiveSourceTimeoutMins,
		DeactivatorIntervalMins:        config.Bot.DeactivatorIntervalMins,
		Port:                           config.Bot.Port,
	}

	if err := bot.RunBot(db, config.Bot.Token, botConfig); err != nil {
		panic(err)
	}
}

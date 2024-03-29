package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot"
	"github.com/aigic8/warmlight/pkg/utils"
)

const DEFAULT_CONFIG_PATH = "warmlight.toml"

func main() {
	var rawConfigPath string
	flag.StringVar(&rawConfigPath, "c", DEFAULT_CONFIG_PATH, fmt.Sprintf("config file path, default value is '%s'", DEFAULT_CONFIG_PATH))
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var configPath string
	if strings.HasPrefix(rawConfigPath, "/") { // path is absolute
		configPath = rawConfigPath
	} else { // path is relative
		configPath = path.Join(cwd, rawConfigPath)
	}

	config, err := utils.LoadConfig(configPath)
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

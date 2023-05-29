package utils

import (
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml/v2"
)

type (
	Config struct {
		Db  *DBConfig  `toml:"db" validate:"required"`
		Bot *BotConfig `toml:"bot" validate:"required"`
	}

	DBConfig struct {
		URL       string `toml:"url" validate:"required"`
		TimeoutMs int    `toml:"timeoutMs" validate:"gte=0"`
	}

	BotConfig struct {
		Token                          string `toml:"token" validate:"required"`
		WebhookURL                     string `toml:"webhookUrl" validate:"required"`
		CertFilePath                   string `toml:"certFile" validate:"required"`
		PrivKeyFilePath                string `toml:"privKeyFile" validate:"required"`
		LogPath                        string `toml:"logPath" validate:"required"`
		IsDev                          bool   `toml:"isDev"`
		DefaultActiveSourceTimeoutMins int    `toml:"defaultActiveSourceTimeoutMins" validate:"gte=0"`
		DeactivatorIntervalMins        int    `toml:"deactivatorIntervalMins" validate:"gte=0"`
		Port                           int    `toml:"port" validate:"gte=0"`
		LibraryTokenExpireMins         int    `toml:"libraryTokenExpireMins" validate:"gte=0"`
	}
)

const DEFAULT_ACTIVE_SOURCE_TIMEOUT = 60
const DEFAULT_DEACTIVATOR_INTERVAL_MINS = 10
const DEFAULT_DB_TIMEOUT_MS = 5000
const DEFAULT_LIBRARY_TOKEN_EXPIRE_MINS = 30
const DEFAULT_PORT = 443

func LoadConfig(configPath string) (*Config, error) {
	// TODO test LoadConfig
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tomlBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := toml.Unmarshal(tomlBytes, &config); err != nil {
		panic(err)
	}

	v := validator.New()
	if err := v.Struct(&config); err != nil {
		return nil, err
	}

	if config.Bot.DefaultActiveSourceTimeoutMins == 0 {
		config.Bot.DefaultActiveSourceTimeoutMins = DEFAULT_ACTIVE_SOURCE_TIMEOUT
	}

	if config.Bot.DeactivatorIntervalMins == 0 {
		config.Bot.DeactivatorIntervalMins = DEFAULT_DEACTIVATOR_INTERVAL_MINS
	}

	if config.Bot.LibraryTokenExpireMins == 0 {
		config.Bot.LibraryTokenExpireMins = DEFAULT_LIBRARY_TOKEN_EXPIRE_MINS
	}

	if config.Db.TimeoutMs == 0 {
		config.Db.TimeoutMs = DEFAULT_DB_TIMEOUT_MS
	}

	if config.Bot.Port == 0 {
		config.Bot.Port = DEFAULT_PORT
	}

	return &config, nil
}

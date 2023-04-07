package main

import (
	"time"

	"github.com/aigic8/warmlight/bot"
	"github.com/aigic8/warmlight/db"
)

// FIXME use configuration file
const DB_URL = "postgresql://postgres:postgres@localhost:1616/warmlight_test"
const DB_TIMEOUT = 5 * time.Second
const TELEGRAM_TOKEN = ""

func main() {
	db, err := db.NewDB(DB_URL, DB_TIMEOUT)
	if err != nil {
		panic(err)
	}

	if err := bot.RunBot(db, TELEGRAM_TOKEN); err != nil {
		panic(err)
	}
}

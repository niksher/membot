package bot

import (
	"database/sql"
	"tg-video-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	API             *tgbotapi.BotAPI
	DB              *sql.DB
	VideoRepository database.VideoRepository
}

func Start(token string, db *sql.DB) error {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	bot := &Bot{
		API:             botAPI,
		DB:              db,
		VideoRepository: *database.NewVideoRepository(db),
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := botAPI.GetUpdatesChan(u)

	for update := range updates {
		bot.HandleUpdate(update)
	}

	return nil
}

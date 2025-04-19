package main

import (
	"log"
	"os"
	"tg-video-bot/internal/bot"
	"tg-video-bot/internal/database"
)

func main() {
	// Инициализация БД
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer db.Close()

	// Запуск бота
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if err := bot.Start(botToken, db); err != nil {
		log.Fatal("Bot failed:", err)
	}
}

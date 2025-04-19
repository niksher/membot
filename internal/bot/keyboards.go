package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

func (b *Bot) ShowMainMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = mainMenuKeyboard()
	b.API.Send(msg)
}

func mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/get_video"),
		),
		/*tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏷 Добавить теги"),
			tgbotapi.NewKeyboardButton("🔍 Найти по тегу"),
		),*/
	)
}

func (b *Bot) ShowAdminMenu(chatID int64) {
	if !b.IsAdmin(chatID) {
		return
	}

	buttons := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/list_videos"),
			tgbotapi.NewKeyboardButton("/add_video"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/delete_video"),
			tgbotapi.NewKeyboardButton("/stats"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "⚙️ Админ-панель")
	msg.ReplyMarkup = buttons
	b.API.Send(msg)
}

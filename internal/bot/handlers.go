package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tg-video-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUpdate обрабатывает все входящие апдейты
func (b *Bot) HandleUpdate(update tgbotapi.Update) {
	switch {
	case update.CallbackQuery != nil:
		b.HandleCallbackQuery(update.CallbackQuery)

	case update.Message != nil:
		if update.Message.IsCommand() {
			b.HandleCommand(update.Message)
		} else if update.Message.Video != nil {
			b.HandleVideoMessage(update.Message)
		} else {
			b.HandleTextMessage(update.Message)
		}
	}
}

// HandleCommand обрабатывает текстовые команды
func (b *Bot) HandleCommand(msg *tgbotapi.Message) {
	s, _ := json.Marshal(msg)
	fmt.Println(string(s))
	switch msg.Command() {
	case "start":
		b.ShowMainMenu(msg.Chat.ID)
		if b.IsAdmin(int64(msg.From.ID)) && msg.Text == "⚙️ Админ-панель" {
			b.ShowAdminMenu(msg.Chat.ID)
		}
	case "help":
		b.SendHelpMessage(msg.Chat.ID)
	case "add_tags":
		b.HandleAddTagsCommand(msg)
	case "get_by_tag":
		b.HandleGetByTagCommand(msg)
	case "get_video":
		b.HandleGetVideoCommand(msg)
	case "get_videos", "get_videos_5":
		b.HandleGetVideosCommand(msg)
	case "add_video":
		b.HandleAddVideoCommand(msg)
	case "list_videos":
		b.HandleListVideosCommand(msg)
	case "delete_video":
		b.HandleDeleteVideoCommand(msg)
	default:
		b.SendUnknownCommand(msg.Chat.ID)
	}
}

// HandleVideoMessage обрабатывает получение видео
func (b *Bot) HandleVideoMessage(msg *tgbotapi.Message) {
	// Проверяем права администратора
	if !b.IsAdmin(int64(msg.From.ID)) {
		//b.SendMessage(msg.Chat.ID, "❌ Только администраторы могут добавлять видео")
		return
	}
	if !b.IsAdminGroup(msg.Chat.ID) {
		//b.SendMessage(msg.Chat.ID, "❌ Только администраторы могут добавлять видео")
		return
	}

	video := models.Video{
		FileID:  msg.Video.FileID,
		Caption: msg.Caption,
	}

	videoID, err := b.VideoRepository.SaveVideo(video)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			b.SendMessage(msg.Chat.ID, "⚠️ Это видео уже есть в базе")
			return
		}
		log.Printf("Ошибка сохранения видео: %v", err)
		b.SendMessage(msg.Chat.ID, "❌ Ошибка сохранения видео")
		return
	}

	response := fmt.Sprintf("✅ Видео сохранено (ID: %d)\nДобавьте теги командой:\n/add_tags %d тег1 тег2", videoID, videoID)
	b.SendMessage(msg.Chat.ID, response)
}

// HandleTextMessage обрабатывает обычные текстовые сообщения
func (b *Bot) HandleTextMessage(msg *tgbotapi.Message) {
	switch msg.Text {
	case "📥 Добавить видео":
		b.SendMessage(msg.Chat.ID, "Отправьте мне видео для сохранения")
	case "🏷 Добавить теги":
		b.SendMessage(msg.Chat.ID, "Введите ID видео и теги через пробел:\nПример: 5 котики смешные")
		/*case "🔍 Найти по тегу":
			b.ShowPopularTags(msg.Chat.ID)
		default:
			b.TryParseAsTagCommand(msg)*/
	}
}

// HandleCallbackQuery обрабатывает нажатия инлайн-кнопок
func (b *Bot) HandleCallbackQuery(query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data

	switch {
	case strings.HasPrefix(data, "tag_"):
		tag := strings.TrimPrefix(data, "tag_")
		b.SendVideosByTag(chatID, tag)

	/*case data == "more_tags":
	b.ShowMoreTags(chatID)*/

	case strings.HasPrefix(data, "video_"):
		videoID, _ := strconv.Atoi(strings.TrimPrefix(data, "video_"))
		b.SendVideoByID(chatID, int64(videoID))
	}

	b.API.AnswerCallbackQuery(tgbotapi.NewCallback(query.ID, ""))
}

// HandleAddTagsCommand обрабатывает команду добавления тегов
func (b *Bot) HandleAddTagsCommand(msg *tgbotapi.Message) {
	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		b.SendMessage(msg.Chat.ID, "Используйте: /add_tags [ID видео] [теги через пробел]")
		return
	}

	videoID, err := strconv.Atoi(args[0])
	if err != nil {
		b.SendMessage(msg.Chat.ID, "❌ Неверный ID видео")
		return
	}

	tags := args[1:]
	err = b.VideoRepository.AddTagsToVideo(int64(videoID), tags)
	if err != nil {
		log.Printf("Ошибка добавления тегов: %v", err)
		b.SendMessage(msg.Chat.ID, "❌ Ошибка добавления тегов")
		return
	}

	b.SendMessage(msg.Chat.ID, fmt.Sprintf("✅ Добавлены теги: %s", strings.Join(tags, ", ")))
}

// HandleGetByTagCommand обрабатывает поиск по тегу
func (b *Bot) HandleGetByTagCommand(msg *tgbotapi.Message) {
	tag := msg.CommandArguments()
	/*if tag == "" {
		b.ShowPopularTags(msg.Chat.ID)
		return
	}*/

	b.SendVideosByTag(msg.Chat.ID, tag)
}

// HandleGetVideoCommand обрабатывает команду /get_video
func (b *Bot) HandleGetVideoCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// Получаем случайное непросмотренное видео
	video, err := b.VideoRepository.GetRandomUnsentVideo(chatID, 1)
	if err != nil {
		if err.Error() == "no unsent videos available" {
			b.SendMessage(chatID, "🎉 Вы уже просмотрели все доступные видео!")
			return
		}
		log.Printf("Failed to get random video: %v", err)
		b.SendMessage(chatID, "❌ Произошла ошибка при получении видео")
		return
	}

	// Отправляем видео
	videoMsg := tgbotapi.NewVideoShare(chatID, video[0].FileID)
	if video[0].Caption != "" {
		videoMsg.Caption = video[0].Caption
	}

	// Добавляем кнопки с тегами
	if len(video[0].Tags) > 0 {
		videoMsg.ReplyMarkup = createVideoTagsKeyboard(video[0].Tags)
	}

	if _, err := b.API.Send(videoMsg); err != nil {
		log.Printf("Failed to send video: %v", err)
		b.SendMessage(chatID, "❌ Не удалось отправить видео")
		return
	}

	// Помечаем видео как отправленное
	if err := b.VideoRepository.MarkVideoSent(chatID, video[0].ID); err != nil {
		log.Printf("Failed to mark video as sent: %v", err)
	}
}

// HandleGetVideosCommand обрабатывает команду /get_videos
func (b *Bot) HandleGetVideosCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	// Получаем случайное непросмотренное видео (по умолчанию 5)
	nums := 5
	args := strings.TrimSpace(msg.CommandArguments())
	if args != "" {
		if parsedNums, err := strconv.Atoi(args); err == nil && parsedNums > 0 {
			nums = parsedNums
		}
	}

	videos, err := b.VideoRepository.GetRandomUnsentVideo(chatID, nums)
	if err != nil {
		if err.Error() == "no unsent videos available" {
			b.SendMessage(chatID, "🎉 Вы уже просмотрели все доступные видео!")
			return
		}
		log.Printf("Failed to get random video: %v", err)
		b.SendMessage(chatID, "❌ Произошла ошибка при получении видео")
		return
	}

	for _, v := range videos {
		// Отправляем видео
		videoMsg := tgbotapi.NewVideoShare(chatID, v.FileID)
		if v.Caption != "" {
			videoMsg.Caption = v.Caption
		}
		if _, err := b.API.Send(videoMsg); err != nil {
			log.Printf("Failed to send video: %v", err)
			b.SendMessage(chatID, "❌ Не удалось отправить видео")
			return
		}
		// Помечаем видео как отправленное
		if err := b.VideoRepository.MarkVideoSent(chatID, v.ID); err != nil {
			log.Printf("Failed to mark video as sent: %v", err)
		}
	}
}

func (b *Bot) HandleAddVideoCommand(msg *tgbotapi.Message) {
	if !b.IsAdmin(int64(msg.From.ID)) {
		//b.SendMessage(msg.Chat.ID, "❌ Недостаточно прав")
		return
	}
	b.SendMessage(msg.Chat.ID, "Отправьте видео для добавления в базу")
}

func (b *Bot) HandleListVideosCommand(msg *tgbotapi.Message) {
	if !b.IsAdmin(int64(msg.From.ID)) {
		b.SendMessage(msg.Chat.ID, "❌ Недостаточно прав")
		return
	}

	videos, err := b.VideoRepository.GetAllVideos()
	if err != nil {
		b.SendMessage(msg.Chat.ID, "❌ Ошибка получения списка видео")
		return
	}

	var response strings.Builder
	response.WriteString("📋 Список видео:\n\n")
	for _, v := range videos {
		response.WriteString(fmt.Sprintf("ID: %d\n", v.ID))
		if v.Caption != "" {
			response.WriteString(fmt.Sprintf("Описание: %s\n", v.Caption))
		}
		response.WriteString("\n")
	}

	b.SendMessage(msg.Chat.ID, response.String())
}

func (b *Bot) HandleDeleteVideoCommand(msg *tgbotapi.Message) {
	if !b.IsAdmin(int64(msg.From.ID)) {
		b.SendMessage(msg.Chat.ID, "❌ Недостаточно прав")
		return
	}

	videoID, err := strconv.Atoi(msg.CommandArguments())
	if err != nil {
		b.SendMessage(msg.Chat.ID, "Используйте: /delete_video [ID видео]")
		return
	}

	if err := b.VideoRepository.DeleteVideo(int64(videoID)); err != nil {
		b.SendMessage(msg.Chat.ID, "❌ Ошибка удаления видео")
		return
	}

	b.SendMessage(msg.Chat.ID, fmt.Sprintf("✅ Видео ID %d удалено", videoID))
}

// SendVideosByTag отправляет видео по указанному тегу
func (b *Bot) SendVideosByTag(chatID int64, tag string) {
	videos, err := b.VideoRepository.GetVideosByTag(tag)
	if err != nil || len(videos) == 0 {
		b.SendMessage(chatID, "❌ Видео с тегом '"+tag+"' не найдено")
		return
	}

	// Отправляем первое видео
	video := videos[0]
	msg := tgbotapi.NewVideoShare(chatID, video.FileID)
	if video.Caption != "" {
		msg.Caption = video.Caption
	}
	b.API.Send(msg)

	// Если есть еще видео - предлагаем кнопку "Показать еще"
	/*if len(videos) > 1 {
		b.ShowMoreVideosButton(chatID, tag, 1)
	}*/
}

// SendVideoByID отправляет конкретное видео по ID
func (b *Bot) SendVideoByID(chatID, videoID int64) {
	video, err := b.VideoRepository.GetVideoByID(videoID)
	if err != nil {
		b.SendMessage(chatID, "❌ Видео не найдено"+err.Error())
		return
	}

	// Проверяем, не отправлялось ли уже это видео
	if b.VideoRepository.IsVideoSent(chatID, videoID) {
		b.SendMessage(chatID, "⚠️ Вы уже получали это видео ранее")
		return
	}

	msg := tgbotapi.NewVideoShare(chatID, video.FileID)
	if video.Caption != "" {
		msg.Caption = video.Caption
	}
	b.API.Send(msg)

	// Запоминаем факт отправки
	b.VideoRepository.MarkVideoSent(chatID, videoID)
}

// Вспомогательные методы для отправки сообщений
func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.API.Send(msg)
}

func (b *Bot) SendHelpMessage(chatID int64) {
	helpText := `📚 Доступные команды:
/get_video - Получить случайное видео
/get_videos [кол-во] - Получить случайные видео (по умолчанию 5)
/add_tags [ID] [теги] - Добавить теги к видео
/get_by_tag [тег] - Найти видео по тегу`
	b.SendMessage(chatID, helpText)
}

func (b *Bot) SendUnknownCommand(chatID int64) {
	b.SendMessage(chatID, "❌ Неизвестная команда. Введите /help для списка команд")
}

// Создает клавиатуру с тегами видео
func createVideoTagsKeyboard(tags []string) tgbotapi.InlineKeyboardMarkup {
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, tag := range tags {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"#"+tag,
				"tag_"+tag,
			),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

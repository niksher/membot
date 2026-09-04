package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// BotCommand представляет команду бота для Telegram
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// SetMyCommandsRequest представляет тело запроса к setMyCommands
type SetMyCommandsRequest struct {
	Commands []BotCommand `json:"commands"`
}

// SetMyCommandsResponse представляет ответ Telegram API
type SetMyCommandsResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// SetCommands регистрирует команды бота в Telegram API (появляются при вводе "/")
func (b *Bot) SetCommands() error {
	commands := []BotCommand{
		{
			Command:     "get_video",
			Description: "Получить случайное видео",
		},
		{
			Command:     "get_videos",
			Description: "Получить 5 случайных видео",
		},
	}

	reqBody, err := json.Marshal(SetMyCommandsRequest{
		Commands: commands,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal bot commands: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", b.API.Token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send setMyCommands request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read setMyCommands response: %w", err)
	}

	var apiResp SetMyCommandsResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return fmt.Errorf("failed to parse setMyCommands response: %w", err)
	}

	if !apiResp.Ok {
		return fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	log.Println("Bot commands registered successfully in Telegram (/get_video, /get_videos)")
	return nil
}

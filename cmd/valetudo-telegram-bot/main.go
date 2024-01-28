package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/bot"
	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type BotConfig struct {
	TelegramBotToken string
	TelegramChatIds  []string
	ValetudoUrl      string
	TelegramDebug    bool
}

func parseTelegramChatIds(chatIds string) []string {
	if chatIds == "" {
		return []string{}
	}

	return strings.Split(chatIds, ",")
}

func loadConfig() *BotConfig {
	return &BotConfig{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatIds:  parseTelegramChatIds(os.Getenv("TELEGRAM_CHAT_IDS")),
		TelegramDebug:    os.Getenv("TELEGRAM_DEBUG") == "true",
		ValetudoUrl:      os.Getenv("VALETUDO_URL"),
	}
}

func main() {
	godotenv.Load(".env")
	config := loadConfig()

	if config.TelegramBotToken == "" {
		log.Panic("TELEGRAM_BOT_TOKEN is not set")
	}

	if config.ValetudoUrl == "" {
		log.Panic("VALETUDO_URL is not set")
	}

	log.Println("Starting Valetudo Telegram Bot")

	api := valetudo.Init(config.ValetudoUrl)
	telegramBot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Panic(fmt.Errorf("failed to initialize telegram integration, have you set your bot token? %w", err))
	}

	telegramBot.Debug = config.TelegramDebug

	botApp := bot.NewBot(&api, telegramBot)

	for _, id := range config.TelegramChatIds {
		chatId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Panic(fmt.Errorf("failed to parse telegram chat id: %w", err))
		}

		botApp.AddUserId(chatId)
	}

	err = botApp.Start()
	if err != nil {
		log.Panic(err)
	}
}

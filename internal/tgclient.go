package internal

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-db-backup-to-s3/internal/types"
	"log"
)

// TgClient клиент для телеграм бота
type TgClient struct {
	config *types.Telegram
	bot    *tgbotapi.BotAPI
}

// NewTgClient конструктор
func NewTgClient(
	config *types.Telegram,
) *TgClient {
	client := &TgClient{
		config: config,
	}
	err := client.initBot()
	if err != nil {
		fmt.Printf("error on tg bot init: %s\n", err)
	}
	return client
}

// initBot инициализация клиента
func (c *TgClient) initBot() error {
	bot, err := tgbotapi.NewBotAPI(c.config.ApiToken)
	if err != nil {
		log.Panic(err)
	}
	c.bot = bot
	return err
}

// SendDbBackupFinishMessage отправка сообщения о завершении бекапа БД
func (c *TgClient) SendDbBackupFinishMessage(dbName, preSignLink string) {
	for _, chatId := range c.config.ChatIds {
		msg := tgbotapi.NewMessage(chatId, "DB backup of "+dbName+" finished\n\nDownload file:\n"+preSignLink)
		_, _ = c.bot.Send(msg)
	}
}

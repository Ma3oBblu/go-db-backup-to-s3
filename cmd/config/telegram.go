package config

// Telegram конфиг для телеграм бота
type Telegram struct {
	ApiToken string
	ChatIds  []int64
}

// NewTelegramConfig конструктор
func NewTelegramConfig(apiToken string, chatIds []int64) *Telegram {
	return &Telegram{
		ApiToken: apiToken,
		ChatIds:  chatIds,
	}
}

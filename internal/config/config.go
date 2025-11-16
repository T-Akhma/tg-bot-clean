package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken      string
	EncryptionKey string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		return Config{}, fmt.Errorf("ENCRYPTION_KEY не задан в .env")
	}
	if l := len(encKey); l != 16 && l != 24 && l != 32 {
		return Config{}, fmt.Errorf("ENCRYPTION_KEY должен быть длиной 16, 24 или 32 символа, сейчас: %d", l)
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN пуст. Проверь .env в корне проекта")
	}

	cfg := Config{
		BotToken:      token,
		EncryptionKey: encKey,
	}
	return cfg, nil
}

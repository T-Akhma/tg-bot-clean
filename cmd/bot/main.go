package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/T-Akhma/tg-bot-clean/internal/config"
	"github.com/T-Akhma/tg-bot-clean/internal/handlers"
	"github.com/T-Akhma/tg-bot-clean/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации:", "err", err)
		return
	}
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		slog.Error("Ошибка создания бота:", "err", err)
		return
	}
	slog.Info("Бот авторизован",
		"username", bot.Self.UserName,
		"securepaste_base_url", cfg.SecurePasteBaseURL,
	)
	encryptionKey := []byte(cfg.EncryptionKey)
	if l := len(encryptionKey); l != 16 && l != 24 && l != 32 {
		slog.Error("Некорректная длина ключа шифрования", "len", l)
		os.Exit(1)
	}
	store := storage.NewFileStore("data_json", encryptionKey)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("⏹ Получен сигнал завершения. Останавливаю polling…")
			bot.StopReceivingUpdates()
			slog.Info("✅ Бот завершил работу.")
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			msg := update.Message
			chatID := msg.Chat.ID

			slog.Info("📩 Новое сообщение",
				"from", msg.From.UserName,
				"chat_id", chatID,
				"is_command", msg.IsCommand(),
				"len", len(msg.Text),
			)

			if msg.IsCommand() {
				handlers.HandleCommand(bot, store, cfg.SecurePasteBaseURL, msg)
				continue
			}
			handlers.SendText(bot, chatID, "Эхо: "+msg.Text)
		}
	}
}

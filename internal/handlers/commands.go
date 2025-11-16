package handlers

import (
	"log/slog"
	"strings"

	"github.com/T-Akhma/tg-bot-clean/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommand(bot *tgbotapi.BotAPI, store *storage.FileStore, msg *tgbotapi.Message) bool {
	chatID := msg.Chat.ID
	cmd := msg.Command()
	switch cmd {
	case "start":
		SendText(bot, chatID, "👋 Привет! Я жив и готов к работе.")

	case "ping":
		SendText(bot, chatID, "pong 🏓")

	case "help":
		SendText(bot, chatID, helpText())

	case "remember":
		args := strings.TrimSpace(msg.CommandArguments())
		if args == "" {
			SendText(bot, chatID, "Используй: /remember текст, который нужно запомнить.")
			return true
		}

		if err := store.Set(chatID, "remembered", args); err != nil {
			slog.Error("Не удалось сохранить заметку в FileStore", "err", err)
			SendText(bot, chatID, "⚠️ Не удалось сохранить заметку, попробуй позже.")
			return true
		}

		SendText(bot, chatID, "💾 Запомнил: "+args)

	case "last":
		if val, ok := store.Get(chatID, "remembered"); ok {
			SendText(bot, chatID, "🧠 Я помню: "+val)
		} else {
			SendText(bot, chatID, "Я ещё ничего не запомнил. Используй /remember.")
		}

	case "clear":
		if err := store.Delete(chatID, "remembered"); err != nil {
			slog.Error("Не удалось удалить заметку из FileStore", "err", err)
			SendText(bot, chatID, "⚠️ Не удалось удалить заметку, попробуй позже.")
			return true
		}
		SendText(bot, chatID, "🧹 Я всё забыл, что ты просил запомнить.")

	default:
		SendText(bot, chatID, "Неизвестная команда. Попробуй /help")
	}

	return true
}

func SendText(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(msg); err != nil {
		slog.Error("Ошибка отправки сообщения", "err", err)
	}
}

func helpText() string {
	return "📜 Доступные команды:\n" +
		"/start    - проверить, что бот жив\n" +
		"/ping     - ответ pong 🏓\n" +
		"/help     - показать это сообщение\n" +
		"/remember - запомнить указанный текст\n" +
		"/last     - показать последний запомненный текст\n" +
		"/clear    - забыть запомненный текст"
}

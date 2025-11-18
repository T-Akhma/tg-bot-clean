package handlers

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/T-Akhma/tg-bot-clean/internal/securepaste"
	"github.com/T-Akhma/tg-bot-clean/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCommand(bot *tgbotapi.BotAPI, store *storage.FileStore, secureBaseURL string, msg *tgbotapi.Message) bool {
	chatID := msg.Chat.ID
	cmd := msg.Command()

	ttlSeconds := 600

	if strings.HasPrefix(cmd, "spaste") {
		suffix := cmd[len("spaste"):]

		if suffix != "" {
			parsedTTL, ok := parseTTLFromSuffix(suffix)
			if !ok {
				SendText(bot, chatID, "❗ Неверный формат команды. Примеры: /spaste, /spaste10m, /spaste30s, /spaste2h")
				return true
			}
			ttlSeconds = parsedTTL
		}
		cmd = "spaste"
	}

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

	case "spaste":
		args := strings.TrimSpace(msg.CommandArguments())
		if args == "" {
			SendText(bot, chatID, "❗ Используй: /spaste текст или /spaste10m текст")
			return true
		}

		url, err := securepaste.CreatePaste(secureBaseURL, ttlSeconds, args)
		if err != nil {
			slog.Error("Не удалось создать пасту в SecurePaste", "err", err)
			SendText(bot, chatID, "⚠️ Не удалось сохранить секрет в SecurePaste, попробуй позже.")
			return true
		}

		text := "🔐 Секрет сохранён в SecurePaste:\n" + url
		SendText(bot, chatID, text)

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
		"/start       - проверить, что бот жив\n" +
		"/ping        - ответ pong 🏓\n" +
		"/help        - показать это сообщение\n" +
		"/remember    - запомнить указанный текст (локально в боте)\n" +
		"/last        - показать последний запомненный текст\n" +
		"/clear       - забыть запомненный текст\n" +
		"/spaste      - сохранить секрет через SecurePaste (TTL по умолчанию: 10m)\n" +
		"/spaste10m   - сохранить секрет на 10 минут\n" +
		"/spaste30s   - сохранить секрет на 30 секунд\n" +
		"/spaste2h    - сохранить секрет на 2 часа"
}

func parseTTLFromSuffix(suffix string) (int, bool) {
	if len(suffix) < 2 {
		return 0, false
	}

	unit := suffix[len(suffix)-1]
	numberPart := suffix[:len(suffix)-1]

	n, err := strconv.Atoi(numberPart)
	if err != nil || n <= 0 {
		return 0, false
	}

	switch unit {
	case 's':
		return n, true
	case 'm':
		return n * 60, true
	case 'h':
		return n * 60 * 60, true
	case 'd':
		return n * 60 * 60 * 24, true
	default:
		return 0, false
	}
}

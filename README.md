# Secure Notes Telegram Bot

Telegram-бот на Go для безопасного хранения заметок прямо в чате.  
Заметки шифруются с помощью AES-GCM и сохраняются в файле, так что после перезапуска бота данные не теряются.

Репозиторий: https://github.com/T-Akhma/tg-bot-clean  

## Integration with SecurePaste

This bot can also work as a client for the [SecurePaste](https://github.com/T-Akhma/securepaste) service.

When `SECUREPASTE_BASE_URL` is configured and the SecurePaste server is running, you can use:

- `/spaste <text>` — send a secret to SecurePaste with default TTL (e.g. 10 minutes).
- `/spaste30s <text>` — secret lives for 30 seconds.
- `/spaste10m <text>` — secret lives for 10 minutes.
- `/spaste2h <text>` — secret lives for 2 hours.

The bot sends the secret to SecurePaste via HTTP (`POST /api/pastes`), which:
- encrypts the content with AES and stores it in PostgreSQL,
- returns an ID,
- and the bot responds with a direct API URL to the created paste:
  `http://127.0.0.1:8080/api/pastes/<id>`.

## Технологии

- Go (Golang)
- Telegram Bot API через библиотеку `github.com/go-telegram-bot-api/telegram-bot-api/v5`
- Структурированные логи: `log/slog`
- Работа с конфигурацией и `.env`
- Собственное файловое хранилище (`FileStore`) поверх JSON
- Шифрование:
  - AES-GCM (`crypto/aes`, `crypto/cipher`)
  - криптографически безопасный рандом (`crypto/rand`)
  - base64 (`encoding/base64`)


## Требования

- Go 1.21+ (или совместимая версия)
- Аккаунт в Telegram и созданный бот через @BotFather (https://t.me/BotFather)

## Установка

    git clone https://github.com/T-Akhma/tg-bot-clean.git
    cd tg-bot-clean

## Настройка окружения

Создайте файл `.env` в корне проекта (рядом с `go.mod`):

    TELEGRAM_BOT_TOKEN=твой_токен_бота_от_BotFather
    ENCRYPTION_KEY=0123456789ABCDEF0123456789ABCDEF

- `TELEGRAM_BOT_TOKEN` — токен, который выдаёт @BotFather.  
- `ENCRYPTION_KEY` — ключ шифрования (строка длиной 16, 24 или 32 символа).  
  Для AES-256 используется ключ длиной **32 байта**, как в примере.

⚠️ Файл `.env` должен быть добавлен в `.gitignore` и **не должен** попадать в репозиторий.

## Запуск

    go run ./cmd/bot

После запуска бот начнёт long-polling обновления от Telegram.

## Пример использования

### Сохранение заметки

Отправьте боту:

    /remember Мой супер секретный текст

Ответ:

    💾 Запомнил: Мой супер секретный текст

### Получение последней заметки

Отправьте:

    /last

Ответ:

    🧠 Я помню: Мой супер секретный текст

### Очистка

Отправьте:

    /clear

Ответ:

    🧹 Я всё забыл, что ты просил запомнить.

После перезапуска бота сохранённые заметки остаются доступными  
(если не менять `ENCRYPTION_KEY` и не удалять `data.json`).

## Безопасность

- Заметки перед записью в файл шифруются c помощью AES-GCM.
- В файле (`data.json`) хранятся **только зашифрованные данные** (base64-строки), а не открытый текст.
- Ключ шифрования (`ENCRYPTION_KEY`) хранится в `.env` и не попадает в репозиторий.
- Структурированные логи не содержат текста заметок — только метаданные (chat_id, длина сообщения, флаги).



## Возможные улучшения

- Поддержка нескольких заметок с ID (`/remember` → выдаёт номер, `/get <id>`).
- Срок жизни заметки (TTL) и авто-удаление.
- Переключение на SQLite-хранилище.
- Дополнительные команды для разработчиков (форматирование JSON, генерация UUID и т.п.).
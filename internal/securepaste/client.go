package securepaste

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type createPasteRequest struct {
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	OneTime bool   `json:"one_time"`
}

type createPasteResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	URL       string    `json:"url,omitempty"`
}

func CreatePaste(baseURL string, ttlSeconds int, text string) (string, error) {

	if strings.TrimSpace(baseURL) == "" {
		return "", fmt.Errorf("CreatePaste: пустой baseURL")
	}

	reqBody := createPasteRequest{
		Content: text,
		TTL:     ttlSeconds,
		OneTime: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("CreatePaste: ошибка JSON-маршалинга: %w", err)
	}

	fullURL := strings.TrimRight(baseURL, "/") + "/api/pastes"

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("CreatePaste: ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("CreatePaste: ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(req.Body, 1024))
		return "", fmt.Errorf("CreatePaste: неожиданный статус %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("CreatePaste: ошибка чтения тела ответа: %w", err)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("CreatePaste: пустое тело ответа при статусе %d", resp.StatusCode)
	}

	var res createPasteResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return "", fmt.Errorf("CreatePaste: ошибка декодирования ответа: %w, data=%s", err, string(data))
	}

	if res.URL != "" {
		return res.URL, nil
	}

	if res.ID != "" {
		pasteURL := strings.TrimRight(baseURL, "/") + "/api/pastes/" + res.ID
		return pasteURL, nil
	}

	return "", fmt.Errorf("CreatePaste: пустой ответ от сервера (нет ни id, ни url)")
}

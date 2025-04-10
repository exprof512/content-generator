package deepseek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/exprof512/content-generator/internal/logger"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    "https://api.deepseek.com/v1",
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) Generate(prompt string) (string, error) {
	reqBody := Request{
		Model: "deepseek-chat",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		logger.Log.WithError(err).Error("Ошибка сериализации тела запроса")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Log.WithError(err).Error("Ошибка создания HTTP-запроса")
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.WithError(err).Error("Ошибка отправки запроса к DeepSeek API")
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			logger.Log.WithFields(map[string]interface{}{
				"status": resp.StatusCode,
			}).WithError(readErr).Error("Ошибка чтения тела ответа от DeepSeek API")
			return "", fmt.Errorf("API error %d occurred, but failed to read response body: %w", resp.StatusCode, readErr)
		}
		logger.Log.WithFields(map[string]interface{}{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("DeepSeek API вернул ошибку")
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Log.WithError(err).Error("Ошибка декодирования ответа от DeepSeek")
		return "", fmt.Errorf("decode error: %w", err)
	}

	if len(response.Choices) == 0 {
		logger.Log.Warn("Ответ DeepSeek пуст — нет Choices")
		return "", fmt.Errorf("empty response")
	}

	logger.Log.WithField("prompt", prompt).Info("Успешная генерация через DeepSeek")
	return response.Choices[0].Message.Content, nil
}

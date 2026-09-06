package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	log        *slog.Logger
}

func CreateClient(baseUrl string, log *slog.Logger) *Client {
	return &Client{baseURL: baseUrl, httpClient: &http.Client{}, log: log}
}

func (c *Client) Download(ctx context.Context, url string) DownloadResponse {
	start := time.Now()
	c.log.Info("worker request started", "url", url)

	payload, err := json.Marshal(DownloadRequest{URL: url})

	if err != nil {
		c.log.Error("marshal request failed", "url", url, "error", err)
		return DownloadResponse{Result: false, Description: fmt.Sprintf("Не удалось сформировать запрос: %v", err)}
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/download", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		c.log.Error("worker request failed", "url", url, "error", err, "took", time.Since(start))
		return DownloadResponse{Result: false, Description: fmt.Sprintf("Не удалось загрузить файл: %v", err)}
	}

	defer resp.Body.Close()

	var result DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.log.Error("decode worker response failed", "url", url, "status", resp.StatusCode, "error", err)
		return DownloadResponse{Result: false, Description: fmt.Sprintf("Не удалось разобрать ответ: %v", err)}
	}

	c.log.Info("worker request finished",
		"url", url,
		"status", resp.StatusCode,
		"result", result.Result,
		"took", time.Since(start))

	return result
}

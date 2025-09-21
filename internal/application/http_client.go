package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/m4ck-y/ETL_go/internal/pkg/logger"
)

type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	if httpErr, ok := err.(*HTTPError); ok {
		switch httpErr.StatusCode {
		case 408, 429, 500, 502, 503, 504:
			return true
		default:
			return false
		}
	}

	return false
}

func retryHTTPRequest(url string, config retryConfig) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= config.maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("User-Agent", "ETL-Service/1.0")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			cancel()
			return resp, nil
		}

		cancel()

		if err != nil {
			lastErr = err
		} else {
			lastErr = &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    resp.Status,
			}
			resp.Body.Close()
		}

		if !isRetryableError(lastErr) {
			break
		}

		if attempt < config.maxRetries {
			delay := time.Duration(float64(config.baseDelay) * math.Pow(2, float64(attempt)))
			if delay > config.maxDelay {
				delay = config.maxDelay
			}

			logger.GlobalLogger.Warn("HTTP request failed, retrying", "system", map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": config.maxRetries + 1,
				"delay":       delay.String(),
				"url":         url,
				"error":       lastErr.Error(),
			})
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", config.maxRetries+1, lastErr)
}

func fetchData(url string, target interface{}, dataType string) error {
	config := retryConfig{
		maxRetries: 3,
		baseDelay:  1 * time.Second,
		maxDelay:   10 * time.Second,
	}

	resp, err := retryHTTPRequest(url, config)
	if err != nil {
		return fmt.Errorf("failed to fetch %s data: %w", dataType, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s response body: %w", dataType, err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse %s JSON: %w", dataType, err)
	}

	return nil
}

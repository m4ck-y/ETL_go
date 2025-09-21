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

// Configuración para reintentos HTTP
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// HTTPError representa un error HTTP con código de estado
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// isRetryableError determina si un error merece reintento
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Errores de red (conexión, DNS, timeout)
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	// Errores HTTP específicos que merecen reintento
	if httpErr, ok := err.(*HTTPError); ok {
		switch httpErr.StatusCode {
		case 408: // Request Timeout
			return true
		case 429: // Too Many Requests
			return true
		case 500: // Internal Server Error
			return true
		case 502: // Bad Gateway
			return true
		case 503: // Service Unavailable
			return true
		case 504: // Gateway Timeout
			return true
		default:
			return false
		}
	}

	return false
}

// retryHTTPRequest realiza petición HTTP con reintentos y backoff exponencial
func retryHTTPRequest(url string, config retryConfig) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= config.maxRetries; attempt++ {
		// Crear contexto con timeout para esta iteración
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Crear petición HTTP
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		// Agregar headers estándar
		req.Header.Set("User-Agent", "ETL-Service/1.0")
		req.Header.Set("Accept", "application/json")

		// Ejecutar petición
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)

		// Verificar respuesta exitosa
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			cancel() // Cancelar contexto antes de retornar
			return resp, nil
		}

		// Cancelar contexto para esta iteración
		cancel()

		// Manejar error
		if err != nil {
			lastErr = err
		} else {
			// Crear error HTTP estructurado
			lastErr = &HTTPError{
				StatusCode: resp.StatusCode,
				Message:    resp.Status,
			}
			resp.Body.Close()
		}

		// Verificar si el error merece reintento
		if !isRetryableError(lastErr) {
			break // No reintentar errores no recuperables
		}

		// Calcular delay para siguiente intento
		if attempt < config.maxRetries {
			delay := time.Duration(float64(config.baseDelay) * math.Pow(2, float64(attempt)))
			if delay > config.maxDelay {
				delay = config.maxDelay
			}

			// Log de reintento con información estructurada
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

// fetchData realiza una petición HTTP con reintentos y parsea la respuesta JSON
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

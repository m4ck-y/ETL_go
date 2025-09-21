package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
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
				"error":       lastErr.Error(),
			})
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("falló después de %d intentos: %w", config.maxRetries+1, lastErr)
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

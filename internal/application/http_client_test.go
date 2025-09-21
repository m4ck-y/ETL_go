package application

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// timeoutError implementa net.Error para testing
type timeoutError struct {
	msg string
}

func (e *timeoutError) Error() string   { return e.msg }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestHTTPError_Error(t *testing.T) {
	err := &HTTPError{
		StatusCode: 404,
		Message:    "Not Found",
	}

	expected := "HTTP 404: Not Found"
	if err.Error() != expected {
		t.Errorf("HTTPError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Network timeout error",
			err:      &timeoutError{msg: "i/o timeout"},
			expected: true,
		},
		{
			name:     "HTTP 408 Request Timeout",
			err:      &HTTPError{StatusCode: 408, Message: "Request Timeout"},
			expected: true,
		},
		{
			name:     "HTTP 429 Too Many Requests",
			err:      &HTTPError{StatusCode: 429, Message: "Too Many Requests"},
			expected: true,
		},
		{
			name:     "HTTP 500 Internal Server Error",
			err:      &HTTPError{StatusCode: 500, Message: "Internal Server Error"},
			expected: true,
		},
		{
			name:     "HTTP 502 Bad Gateway",
			err:      &HTTPError{StatusCode: 502, Message: "Bad Gateway"},
			expected: true,
		},
		{
			name:     "HTTP 503 Service Unavailable",
			err:      &HTTPError{StatusCode: 503, Message: "Service Unavailable"},
			expected: true,
		},
		{
			name:     "HTTP 504 Gateway Timeout",
			err:      &HTTPError{StatusCode: 504, Message: "Gateway Timeout"},
			expected: true,
		},
		{
			name:     "HTTP 400 Bad Request",
			err:      &HTTPError{StatusCode: 400, Message: "Bad Request"},
			expected: false,
		},
		{
			name:     "HTTP 401 Unauthorized",
			err:      &HTTPError{StatusCode: 401, Message: "Unauthorized"},
			expected: false,
		},
		{
			name:     "HTTP 404 Not Found",
			err:      &HTTPError{StatusCode: 404, Message: "Not Found"},
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)

			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestRetryConfig(t *testing.T) {
	config := retryConfig{
		maxRetries: 3,
		baseDelay:  1 * time.Second,
		maxDelay:   10 * time.Second,
	}

	if config.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", config.maxRetries)
	}

	if config.baseDelay != 1*time.Second {
		t.Errorf("Expected baseDelay to be 1s, got %v", config.baseDelay)
	}

	if config.maxDelay != 10*time.Second {
		t.Errorf("Expected maxDelay to be 10s, got %v", config.maxDelay)
	}
}

// Test para verificar configuración de timeout
func TestRetryHTTPRequestTimeout(t *testing.T) {
	// Test más simple que verifica la configuración
	config := retryConfig{
		maxRetries: 2,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   1 * time.Second,
	}

	// URL inexistente que debería fallar rápidamente
	url := "http://nonexistent-domain-that-should-fail-fast.com"

	start := time.Now()
	_, err := retryHTTPRequest(url, config)
	duration := time.Since(start)

	// Debería fallar
	if err == nil {
		t.Error("Expected error for nonexistent domain, but got nil")
	}

	// No debería tomar demasiado tiempo (menos de 5 segundos para este caso)
	if duration > 5*time.Second {
		t.Errorf("Expected duration < 5s for fast failure, got %v", duration)
	}

	// Verificar que el error contiene información útil
	if !strings.Contains(err.Error(), "attempts") {
		t.Errorf("Expected error to contain 'attempts', got: %v", err)
	}
}

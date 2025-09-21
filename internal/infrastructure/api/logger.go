package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel representa el nivel de logging
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	DEBUG LogLevel = "DEBUG"
)

// LogEntry representa una entrada de log estructurada
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Service   string                 `json:"service"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// Logger maneja el logging estructurado
type Logger struct {
	serviceName string
	minLevel    LogLevel
}

// NewLogger crea un nuevo logger
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
		minLevel:    INFO, // Nivel mínimo por defecto
	}
}

// log escribe una entrada de log estructurada
func (l *Logger) log(level LogLevel, message string, requestID string, extra map[string]interface{}) {
	if level < l.minLevel {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		RequestID: requestID,
		Service:   l.serviceName,
		Extra:     extra,
	}

	// Convertir a JSON y escribir
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback a log básico si falla JSON
		log.Printf("[%s] %s: %s", level, l.serviceName, message)
		return
	}

	fmt.Fprintln(os.Stdout, string(jsonData))
}

// Info escribe un log de nivel INFO
func (l *Logger) Info(message string, requestID string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(INFO, message, requestID, extraData)
}

// Error escribe un log de nivel ERROR
func (l *Logger) Error(message string, requestID string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(ERROR, message, requestID, extraData)
}

// Warn escribe un log de nivel WARN
func (l *Logger) Warn(message string, requestID string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(WARN, message, requestID, extraData)
}

// Debug escribe un log de nivel DEBUG
func (l *Logger) Debug(message string, requestID string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(DEBUG, message, requestID, extraData)
}

// Global logger instance
var AppLogger = NewLogger("etl-go-service")

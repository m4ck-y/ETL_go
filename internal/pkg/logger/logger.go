package logger

import (
	"encoding/json"
	"log"
	"time"
)

// LogLevel representa el nivel de logging
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	FATAL LogLevel = "FATAL"
)

// StructuredLog representa una entrada de log estructurada
type StructuredLog struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Service   string                 `json:"service"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// Logger es el logger global del sistema
type Logger struct {
	serviceName string
}

// NewLogger crea un nuevo logger
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
	}
}

// log escribe una entrada de log estructurada en formato JSON
func (l *Logger) log(level LogLevel, message string, requestID string, extra map[string]interface{}) {
	// Crear entrada de log estructurada
	logEntry := StructuredLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Service:   l.serviceName,
		Message:   message,
		RequestID: requestID,
		Extra:     extra,
	}

	// Convertir a JSON y escribir
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		// Fallback si falla el JSON
		log.Printf("[ERROR] %s - Failed to marshal log entry: %v", l.serviceName, err)
		return
	}

	// Escribir JSON estructurado
	log.Println(string(jsonData))
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

// Fatal escribe un log de nivel FATAL y termina el programa
func (l *Logger) Fatal(message string, requestID string, extra ...map[string]interface{}) {
	var extraData map[string]interface{}
	if len(extra) > 0 {
		extraData = extra[0]
	}
	l.log(FATAL, message, requestID, extraData)
	// Para errores fatales, terminamos el programa
	log.Fatal(message)
}

// Logger global
var GlobalLogger = NewLogger("etl-go-service")

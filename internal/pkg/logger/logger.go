package logger

import (
	"log"
)

// LogLevel representa el nivel de logging
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	FATAL LogLevel = "FATAL"
)

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

// log escribe una entrada de log estructurada
func (l *Logger) log(level LogLevel, message string, requestID string, extra map[string]interface{}) {
	// Salida simple estructurada
	log.Printf("[%s] %s - %s (request_id: %s)", level, l.serviceName, message, requestID)
	if extra != nil && len(extra) > 0 {
		log.Printf("Extra: %+v", extra)
	}
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

package api

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/m4ck-y/ETL_go/internal/pkg/logger"
)

// RequestIDMiddleware genera un ID único para cada petición HTTP
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generar ID único para la petición
		requestID := generateRequestID()

		// Agregar al contexto de Gin
		c.Set("request_id", requestID)

		// Agregar header de respuesta para debugging
		c.Header("X-Request-ID", requestID)

		// Log de entrada
		logger.GlobalLogger.Info("Request started", requestID, map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"ip":     c.ClientIP(),
		})

		// Continuar con el siguiente handler
		c.Next()

		// Log de salida con status code
		status := c.Writer.Status()
		if status >= 400 {
			logger.GlobalLogger.Error("Request completed", requestID, map[string]interface{}{
				"status": status,
				"ip":     c.ClientIP(),
			})
		} else {
			logger.GlobalLogger.Info("Request completed", requestID, map[string]interface{}{
				"status": status,
				"ip":     c.ClientIP(),
			})
		}

		if status >= 400 {
			logger.GlobalLogger.Error("Request completed", requestID, map[string]interface{}{
				"status": status,
				"ip":     c.ClientIP(),
			})
		} else {
			logger.GlobalLogger.Info("Request completed", requestID, map[string]interface{}{
				"status": status,
				"ip":     c.ClientIP(),
			})
		}
	}
}

// generateRequestID genera un ID único para la petición
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// GetRequestID obtiene el ID de petición del contexto
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// isBatchAlreadyProcessed verifica si un lote ya fue procesado
func (h *APIHandler) isBatchAlreadyProcessed(c *gin.Context, batchID string) bool {
	requestID := GetRequestID(c)
	processed, err := h.Repo.IsBatchProcessed(batchID)
	if err != nil {
		logger.GlobalLogger.Error("Error verificando estado del lote", requestID, map[string]interface{}{
			"batch_id": batchID,
			"error":    err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check batch status"})
		return true
	}

	if processed {
		logger.GlobalLogger.Info("Lote ya procesado, omitiendo ETL", requestID, map[string]interface{}{
			"batch_id": batchID,
		})
		c.JSON(http.StatusOK, gin.H{"status": "ETL already completed", "batch_id": batchID})
		return true
	}

	return false
}

// markBatchAsProcessed marca un lote como procesado
func (h *APIHandler) markBatchAsProcessed(batchID string) {
	if err := h.Repo.MarkBatchProcessed(batchID); err != nil {
		logger.GlobalLogger.Warn("Error marcando lote como procesado", "system", map[string]interface{}{
			"batch_id": batchID,
			"error":    err.Error(),
		})
		// No fallar el request, solo loggear
	}
}

package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// isBatchAlreadyProcessed verifica si un lote ya fue procesado
func (h *APIHandler) isBatchAlreadyProcessed(c *gin.Context, batchID string) bool {
	processed, err := h.Repo.IsBatchProcessed(batchID)
	if err != nil {
		log.Printf("Error verificando estado del lote: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check batch status"})
		return true
	}

	if processed {
		log.Printf("Lote %s ya procesado, omitiendo ETL", batchID)
		c.JSON(http.StatusOK, gin.H{"status": "ETL already completed", "batch_id": batchID})
		return true
	}

	return false
}

// markBatchAsProcessed marca un lote como procesado
func (h *APIHandler) markBatchAsProcessed(batchID string) {
	if err := h.Repo.MarkBatchProcessed(batchID); err != nil {
		log.Printf("Advertencia: Error marcando lote como procesado: %v", err)
		// No fallar el request, solo loggear
	}
}

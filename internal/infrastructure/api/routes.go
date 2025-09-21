package api

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/m4ck-y/ETL_go/internal/domain"
	"github.com/m4ck-y/ETL_go/internal/domain/models"

	"github.com/m4ck-y/ETL_go/internal/application"
)

type APIHandler struct {
	Repo domain.MetricsRepository
}

// generateBatchID crea un identificador único para lotes ETL
func generateBatchID(adsURL, crmURL, sinceParam string) string {
	// Incluir timestamp diario para granularidad por día
	input := fmt.Sprintf("%s|%s|%s|%d", adsURL, crmURL, sinceParam, time.Now().Unix()/86400)
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)[:16]
}

// parseSinceDate parsea el parámetro opcional de fecha
func parseSinceDate(sinceParam string) (*time.Time, error) {
	if sinceParam == "" {
		return nil, nil
	}

	parsedDate, err := time.Parse("2006-01-02", sinceParam)
	if err != nil {
		return nil, fmt.Errorf("formato de fecha inválido. Use YYYY-MM-DD")
	}

	return &parsedDate, nil
}

// validateEnvironment valida que las variables de entorno necesarias estén configuradas
func validateEnvironment() error {
	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")

	if adsURL == "" || crmURL == "" {
		return fmt.Errorf("ADS_API_URL y CRM_API_URL deben estar configuradas")
	}

	return nil
}

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/ingest/run", h.IngestHandler)
	router.GET("/metrics", h.GetMetricsHandler)
}

// IngestHandler inicia el proceso ETL y guarda los resultados.
// @Summary Ejecuta el proceso ETL de ingestión
// @Description Ejecuta un proceso ETL que extrae datos de ADS y CRM y guarda los resultados. Soporta filtrado por fecha con el parámetro 'since'.
// @Tags ingest
// @Accept json
// @Produce json
// @Param since query string false "Fecha desde la cual filtrar datos (YYYY-MM-DD)"
// @Success 201 {object} map[string]string "ETL completado correctamente"
// @Failure 400 {object} map[string]string "Parámetro de fecha inválido"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /ingest/run [post]
func (h *APIHandler) IngestHandler(c *gin.Context) {
	// Validar configuración
	if err := validateEnvironment(); err != nil {
		log.Printf("Configuración inválida: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")
	log.Printf("Iniciando ETL con URLs configuradas")

	// Parsear parámetro opcional de fecha
	sinceParam := c.Query("since")
	sinceDate, err := parseSinceDate(sinceParam)
	if err != nil {
		log.Printf("Error parseando fecha: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sinceDate != nil {
		log.Printf("Filtrando datos desde: %s", sinceParam)
	}

	// Generar ID único para el lote (idempotencia)
	batchID := generateBatchID(adsURL, crmURL, sinceParam)
	log.Printf("ID de lote generado: %s", batchID)

	// Verificar si el lote ya fue procesado
	if h.isBatchAlreadyProcessed(c, batchID) {
		return
	}

	// Ejecutar proceso ETL
	result, err := application.RunETL(adsURL, crmURL, sinceDate)
	if err != nil {
		log.Printf("Proceso ETL falló: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ETL failed", "details": err.Error()})
		return
	}

	// Guardar resultados
	if err := h.Repo.Save(result); err != nil {
		log.Printf("Error guardando resultados: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save ETL results", "details": err.Error()})
		return
	}

	// Marcar lote como procesado
	h.markBatchAsProcessed(batchID)

	log.Printf("ETL completado exitosamente. %d combinaciones UTM procesadas", len(result))
	c.JSON(http.StatusCreated, gin.H{
		"status":                 "ETL completed",
		"processed_combinations": len(result),
		"batch_id":               batchID,
	})
}

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

// GetMetricsHandler obtiene todas las métricas almacenadas.
// @Summary Obtiene métricas almacenadas
// @Description Retorna un listado de métricas con información de campañas, clics, costo, leads y ventas.
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {array} models.MetricResponse "Lista de métricas"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /metrics [get]
func (h *APIHandler) GetMetricsHandler(c *gin.Context) {
	data, err := h.Repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Transform map to slice for JSON serialization
	var response []models.MetricResponse
	for key, m := range data {
		response = append(response, models.MetricResponse{
			UTMCampaign:   key.Campaign,
			UTMSource:     key.Source,
			UTMMedium:     key.Medium,
			Clicks:        m.Clicks,
			Cost:          m.Cost,
			Leads:         m.Leads,
			Opportunities: m.Opportunities,
			ClosedWon:     m.ClosedWon,
			Revenue:       m.Revenue,
		})
	}

	c.JSON(http.StatusOK, response)
}

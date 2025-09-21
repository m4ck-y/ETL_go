package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/m4ck-y/ETL_go/internal/domain"
	"github.com/m4ck-y/ETL_go/internal/domain/models"
	"github.com/m4ck-y/ETL_go/internal/pkg/logger"

	"github.com/m4ck-y/ETL_go/internal/application"
)

type APIHandler struct {
	Repo domain.MetricsRepository
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
	requestID := GetRequestID(c)

	if err := validateEnvironment(); err != nil {
		logger.GlobalLogger.Error("Configuración inválida", requestID, map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")
	logger.GlobalLogger.Info("Iniciando ETL con URLs configuradas", requestID, map[string]interface{}{
		"ads_url": adsURL,
		"crm_url": crmURL,
	})

	sinceParam := c.Query("since")
	sinceDate, err := parseSinceDate(sinceParam)
	if err != nil {
		logger.GlobalLogger.Error("Error parseando fecha", requestID, map[string]interface{}{
			"since_param": sinceParam,
			"error":       err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sinceDate != nil {
		logger.GlobalLogger.Info("Filtrando datos desde fecha", requestID, map[string]interface{}{
			"since_date": sinceParam,
		})
	}

	batchID := generateBatchID(adsURL, crmURL, sinceParam)
	logger.GlobalLogger.Info("ID de lote generado", requestID, map[string]interface{}{
		"batch_id": batchID,
	})

	if h.isBatchAlreadyProcessed(c, batchID) {
		return
	}

	result, err := application.RunETL(adsURL, crmURL, sinceDate)
	if err != nil {
		logger.GlobalLogger.Error("Proceso ETL falló", requestID, map[string]interface{}{
			"batch_id": batchID,
			"error":    err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ETL failed", "details": err.Error()})
		return
	}

	if err := h.Repo.Save(result); err != nil {
		logger.GlobalLogger.Error("Error guardando resultados", requestID, map[string]interface{}{
			"batch_id": batchID,
			"error":    err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save ETL results", "details": err.Error()})
		return
	}

	h.markBatchAsProcessed(batchID)

	logger.GlobalLogger.Info("ETL completado exitosamente", requestID, map[string]interface{}{
		"batch_id":               batchID,
		"processed_combinations": len(result),
	})

	c.JSON(http.StatusCreated, gin.H{
		"status":                 "ETL completed",
		"processed_combinations": len(result),
		"batch_id":               batchID,
	})
}

// GetMetricsHandler obtiene todas las métricas almacenadas.
// @Summary Obtiene métricas almacenadas con cálculos derivados
// @Description Retorna un listado de métricas con información de campañas, clics, costo, leads, ventas y métricas calculadas (CPC, CPA, CVR, ROAS).
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {array} models.MetricResponse "Lista de métricas con cálculos incluidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /metrics [get]
func (h *APIHandler) GetMetricsHandler(c *gin.Context) {
	requestID := GetRequestID(c)

	data, err := h.Repo.GetAll()
	if err != nil {
		logger.GlobalLogger.Error("Error obteniendo métricas", requestID, map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	var response []models.MetricResponse
	for key, m := range data {
		cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas := application.CalculateDerivedMetrics(m)

		response = append(response, models.MetricResponse{
			Channel:       m.Channel,
			UTMCampaign:   key.Campaign,
			UTMSource:     key.Source,
			UTMMedium:     key.Medium,
			Clicks:        m.Clicks,
			Cost:          m.Cost,
			Leads:         m.Leads,
			Opportunities: m.Opportunities,
			ClosedWon:     m.ClosedWon,
			Revenue:       m.Revenue,
			CPC:           cpc,
			CPA:           cpa,
			CVRLeadToOpp:  cvrLeadToOpp,
			CVROppToWon:   cvrOppToWon,
			ROAS:          roas,
		})
	}

	logger.GlobalLogger.Info("Métricas obtenidas exitosamente", requestID, map[string]interface{}{
		"total_metrics": len(response),
	})

	c.JSON(http.StatusOK, response)
}

// ResetHandler limpia todos los datos almacenados en memoria.
// @Summary Resetea todos los datos almacenados
// @Description Limpia completamente la base de datos en memoria, eliminando todas las métricas y lotes procesados.
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Datos reseteados correctamente"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /admin/reset [post]
func (h *APIHandler) ResetHandler(c *gin.Context) {
	requestID := GetRequestID(c)

	logger.GlobalLogger.Info("Iniciando reset de datos", requestID, nil)

	if err := h.Repo.Clear(); err != nil {
		logger.GlobalLogger.Error("Error reseteando datos", requestID, map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset data", "details": err.Error()})
		return
	}

	logger.GlobalLogger.Info("Datos reseteados exitosamente", requestID, nil)

	c.JSON(http.StatusOK, gin.H{
		"status":  "Data reset successfully",
		"message": "All stored metrics and batch data have been cleared",
	})
}

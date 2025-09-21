package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/m4ck-y/ETL_go/internal/application"
	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// GetChannelMetricsHandler obtiene métricas filtradas por canal
// @Summary Obtiene métricas por canal
// @Description Retorna métricas filtradas por canal con soporte para fechas y paginación
// @Tags metrics
// @Accept json
// @Produce json
// @Param from query string false "Fecha desde (YYYY-MM-DD)"
// @Param to query string false "Fecha hasta (YYYY-MM-DD)"
// @Param channel query string false "Canal específico"
// @Param limit query int false "Límite de resultados" default(50)
// @Param offset query int false "Offset para paginación" default(0)
// @Success 200 {array} models.MetricResponse
// @Failure 500 {object} map[string]string
// @Router /metrics/channel [get]
func (h *APIHandler) GetChannelMetricsHandler(c *gin.Context) {
	// Obtener datos base
	data, err := h.Repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Convertir a slice para procesamiento
	var metrics []models.MetricResponse
	for key, m := range data {
		cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas := application.CalculateDerivedMetrics(m)
		metrics = append(metrics, models.MetricResponse{
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

	// Aplicar filtros
	channel := c.Query("channel")
	metrics = filterMetricsByChannel(metrics, channel)

	// Aplicar paginación
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	metrics = paginateMetrics(metrics, limit, offset)

	c.JSON(http.StatusOK, metrics)
}

// GetFunnelMetricsHandler obtiene métricas de funnel por campaña
// @Summary Obtiene métricas de funnel por campaña
// @Description Retorna métricas de conversión filtradas por campaña con soporte para fechas y paginación
// @Tags metrics
// @Accept json
// @Produce json
// @Param from query string false "Fecha desde (YYYY-MM-DD)"
// @Param to query string false "Fecha hasta (YYYY-MM-DD)"
// @Param utm_campaign query string false "Campaña específica"
// @Param limit query int false "Límite de resultados" default(50)
// @Param offset query int false "Offset para paginación" default(0)
// @Success 200 {array} models.MetricResponse
// @Failure 500 {object} map[string]string
// @Router /metrics/funnel [get]
func (h *APIHandler) GetFunnelMetricsHandler(c *gin.Context) {
	// Obtener datos base
	data, err := h.Repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Convertir a slice para procesamiento
	var metrics []models.MetricResponse
	for key, m := range data {
		cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas := application.CalculateDerivedMetrics(m)
		metrics = append(metrics, models.MetricResponse{
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

	// Aplicar filtros
	campaign := c.Query("utm_campaign")
	metrics = filterMetricsByCampaign(metrics, campaign)

	// Aplicar paginación
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	metrics = paginateMetrics(metrics, limit, offset)

	c.JSON(http.StatusOK, metrics)
}

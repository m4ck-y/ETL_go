package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/m4ck-y/ETL_go/internal/domain"
	"github.com/m4ck-y/ETL_go/internal/domain/models"

	"github.com/m4ck-y/ETL_go/internal/application"
)

type APIHandler struct {
	Repo domain.MetricsRepository
}

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/ingest/run", h.IngestHandler)
	router.GET("/metrics", h.GetMetricsHandler)
}

// IngestHandler inicia el proceso ETL y guarda los resultados.
// @Summary Ejecuta el proceso ETL de ingestión
// @Description Ejecuta un proceso ETL que extrae datos de ADS y CRM y guarda los resultados.
// @Tags ingest
// @Accept json
// @Produce json
// @Success 201 {object} map[string]string "ETL completado correctamente"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /ingest/run [post]
func (h *APIHandler) IngestHandler(c *gin.Context) {
	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")

	log.Println("adsURL:", adsURL)
	log.Println("crmURL:", crmURL)

	result, err := application.RunETL(adsURL, crmURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ETL failed", "details": err.Error()})
		return
	}

	if err := h.Repo.Save(result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save ETL results", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ETL completed"})
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

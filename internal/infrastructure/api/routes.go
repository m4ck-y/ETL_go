package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/m4ck-y/ETL_go/internal/domain"

	"github.com/m4ck-y/ETL_go/internal/application"
)

type APIHandler struct {
	Repo domain.MetricsRepository
}

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/ingest/run", h.IngestHandler)
	router.GET("/metrics", h.GetMetricsHandler)
}

func (h *APIHandler) IngestHandler(c *gin.Context) {
	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")

	log.Println("adsURL:", adsURL)
	fmt.Println("adsURL:", adsURL)
	log.Println("crmURL:", crmURL)
	fmt.Println("crmURL:", crmURL)

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

func (h *APIHandler) GetMetricsHandler(c *gin.Context) {
	data, err := h.Repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}

	// Transform map to slice for JSON serialization
	var response []gin.H
	for key, m := range data {
		response = append(response, gin.H{
			"utm_campaign":  key.Campaign,
			"utm_source":    key.Source,
			"utm_medium":    key.Medium,
			"clicks":        m.Clicks,
			"cost":          m.Cost,
			"leads":         m.Leads,
			"opportunities": m.Opportunities,
			"closed_won":    m.ClosedWon,
			"revenue":       m.Revenue,
		})
	}

	c.JSON(http.StatusOK, response)
}

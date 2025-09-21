package api

import (
	"github.com/gin-gonic/gin"
)

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/ingest/run", h.IngestHandler)
	router.GET("/metrics", h.GetMetricsHandler)
}

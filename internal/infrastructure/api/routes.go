package api

import (
	"github.com/gin-gonic/gin"
)

func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	// Middleware para logging estructurado y request IDs
	router.Use(RequestIDMiddleware())

	router.POST("/ingest/run", h.IngestHandler)
	router.GET("/metrics", h.GetMetricsHandler)
	router.GET("/metrics/channel", h.GetChannelMetricsHandler)
	router.GET("/metrics/funnel", h.GetFunnelMetricsHandler)

	// Health checks
	router.GET("/healthz", h.HealthzHandler)
	router.GET("/readyz", h.ReadyzHandler)

	// Admin endpoints
	router.POST("/admin/reset", h.ResetHandler)
}

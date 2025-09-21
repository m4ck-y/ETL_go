package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse representa la respuesta de health check
type HealthResponse struct {
	Status  string    `json:"status"`
	Time    time.Time `json:"time"`
	Service string    `json:"service"`
}

// ReadinessResponse representa la respuesta de readiness check
type ReadinessResponse struct {
	Status  string            `json:"status"`
	Time    time.Time         `json:"time"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

// HealthzHandler endpoint básico de health check
// @Summary Health check básico
// @Description Verifica que el servicio esté funcionando
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *APIHandler) HealthzHandler(c *gin.Context) {
	response := HealthResponse{
		Status:  "ok",
		Time:    time.Now(),
		Service: "etl-go-service",
	}

	c.JSON(http.StatusOK, response)
}

// ReadyzHandler endpoint de readiness check
// @Summary Readiness check
// @Description Verifica que el servicio esté listo para recibir tráfico
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} ReadinessResponse
// @Failure 503 {object} ReadinessResponse
// @Router /readyz [get]
func (h *APIHandler) ReadyzHandler(c *gin.Context) {
	checks := make(map[string]string)

	// Verificar configuración de entorno
	if err := validateEnvironment(); err != nil {
		checks["environment"] = "failed"
		response := ReadinessResponse{
			Status:  "not ready",
			Time:    time.Now(),
			Service: "etl-go-service",
			Checks:  checks,
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}
	checks["environment"] = "ok"

	// Verificar repositorio
	if h.Repo == nil {
		checks["repository"] = "failed"
		response := ReadinessResponse{
			Status:  "not ready",
			Time:    time.Now(),
			Service: "etl-go-service",
			Checks:  checks,
		}
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}
	checks["repository"] = "ok"

	// Si todas las verificaciones pasan
	response := ReadinessResponse{
		Status:  "ready",
		Time:    time.Now(),
		Service: "etl-go-service",
		Checks:  checks,
	}

	c.JSON(http.StatusOK, response)
}

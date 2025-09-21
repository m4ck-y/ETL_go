package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/api"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/repository"
	"github.com/m4ck-y/ETL_go/internal/pkg/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/m4ck-y/ETL_go/docs" // importa la documentación generada
)

func main() {
	err := godotenv.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			logger.GlobalLogger.Warn("Error al cargar variables de entorno", "system", map[string]interface{}{
				"error": err.Error(),
			})
		}
		logger.GlobalLogger.Info("Variables de entorno cargadas desde Docker", "system", nil)
	} else {
		logger.GlobalLogger.Info("Variables de entorno cargadas desde .env", "system", nil)
	}

	repo := repository.NewInMemoryMetricsRepository()
	handler := &api.APIHandler{Repo: repo}

	router := gin.Default()

	router.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/swagger" || path == "/swagger/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	handler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.GlobalLogger.Info("Servidor escuchando", "system", map[string]interface{}{
		"port": port,
	})
	if err := router.Run(":" + port); err != nil {
		logger.GlobalLogger.Fatal("Error al iniciar el servidor", "system", map[string]interface{}{
			"port":  port,
			"error": err.Error(),
		})
	}
}

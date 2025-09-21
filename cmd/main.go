package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/api"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/repository"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/m4ck-y/ETL_go/docs" // importa la documentación generada
)

func main() {
	//application.Saludar()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error al cargar variables de entorno,")
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

	//api.RegisterRoutes(router)
	handler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[INFO] etl-go-service - Servidor escuchando en puerto: %s (request_id: system)", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

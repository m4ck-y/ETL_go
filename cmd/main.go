package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/api"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/repository"
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

	//api.RegisterRoutes(router)
	handler.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor escuchando en puerto: %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

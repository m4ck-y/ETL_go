package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/m4ck-y/ETL_go/internal/application"
	"github.com/m4ck-y/ETL_go/internal/infrastructure/services"
)

func main() {
	application.Saludar()

	router := gin.Default()

	services.RegisterRoutes(router)

	log.Printf("Servidor escuchando en http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

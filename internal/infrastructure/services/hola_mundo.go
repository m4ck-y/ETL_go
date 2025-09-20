package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registra todas las rutas del servidor
func RegisterRoutes(r *gin.Engine) {
	r.GET("/", helloHandler)
}

// helloHandler maneja el endpoint raíz "/"
func helloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"mensaje": "Hola Mundo",
	})
}

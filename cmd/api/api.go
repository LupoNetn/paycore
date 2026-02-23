package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *Application) SetupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

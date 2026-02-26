package transfer

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/paycore/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler) {
	transferGroup := r.Group("/transfer")
	
	//use middlewares
	transferGroup.Use(middleware.AuthMiddleware())

	//implement routes
	{
		transferGroup.POST("/", h.HandleCreateTransaction)
		transferGroup.GET("/:id", h.HandleGetTransactionByID)
	}
}

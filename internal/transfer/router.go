package transfer

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/paycore/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler, secret string) {
	transferGroup := r.Group("/transfer")
	
	//use middlewares
	transferGroup.Use(middleware.AuthMiddleware(secret))

	//implement routes
	{
		transferGroup.POST("/", h.HandleCreateTransaction)
		transferGroup.GET("/:id", h.HandleGetTransactionByID)
	}
}

package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/paycore/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler) {
	walletGroup := r.Group("/wallets")
	walletGroup.Use(middleware.AuthMiddleware())
	{
		walletGroup.GET("/:id", h.GetWalletHandler)
	}
}
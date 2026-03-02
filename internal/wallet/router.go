package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/luponetn/paycore/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler, secret string) {
	walletGroup := r.Group("/wallets")
	walletGroup.Use(middleware.AuthMiddleware(secret))
	{
		walletGroup.GET("/:id", h.GetWalletHandler)
		walletGroup.GET("/:id/transactions", h.GetWalletTransactionsHandler)
	}
}

package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{Svc: svc}
}

//implement handlers for handling wallet operations.

func (h *Handler) GetWalletHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "wallet id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "user id not found in context"})
		return
	}

	if userID != uuid.MustParse(id) {
		c.JSON(403, gin.H{"error": "forbidden: you can only access your own wallet"})
		return
	}

	wallet, err := h.Svc.GetWalletService(c.Request.Context(), uuid.MustParse(id))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "wallet fetched successfully",
		"wallet":  wallet,
	})
}

// GetWalletTransactionsHandler handles GET /wallet/:id/transactions
func (h *Handler) GetWalletTransactionsHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet id is required"})
		return
	}

	walletID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet id"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id in context is not a valid UUID"})
		return
	}

	if userID != walletID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: you can only access your own wallet"})
		return
	}

	transactions, err := h.Svc.GetWalletTransactionsService(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "transactions fetched successfully",
		"transactions": transactions,
	})
}

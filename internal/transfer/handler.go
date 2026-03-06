package transfer

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreateTransaction(c *gin.Context) {
	userIdVal, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userIdVal.(uuid.UUID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	transaction, err := h.svc.CreateTransaction(c.Request.Context(), userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInsufficientFunds):
			status = http.StatusConflict
		case errors.Is(err, ErrSameWallet), errors.Is(err, ErrInvalidAmount), errors.Is(err, ErrCurrencyMismatch):
			status = http.StatusBadRequest
		case errors.Is(err, ErrWalletNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrUnauthorizedWallet):
			status = http.StatusForbidden
		}

		c.AbortWithStatusJSON(status, gin.H{
			"message": "failed to create transaction",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transaction created successfully",
		"data":    transaction,
	})
}

func (h *Handler) HandleGetTransactionByID(c *gin.Context) {
	transactionID := c.Param("id")
	if transactionID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	transactionIDUUID, err := uuid.Parse(transactionID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	transaction, err := h.svc.GetTransactionByID(c.Request.Context(), transactionIDUUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch transaction",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":     "transaction fetched successfully",
		"transaction": transaction,
	})
}

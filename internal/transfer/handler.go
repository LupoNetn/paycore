package transfer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreateTransaction(c *gin.Context) {
   var req CreateTransactionRequest
   if err := c.ShouldBindJSON(&req); err != nil {
	  c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	  return
   }

   //convert req.Amount to decimal.Decimal and check if it's greater than 0
   amount := decimal.NewFromBigInt(req.Amount.Int, req.Amount.Exp)
   if amount.LessThanOrEqual(decimal.Zero) {
	  c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
	  return
   }
  
   transaction, err := h.svc.CreateTransaction(c.Request.Context(), req)
   if err != nil {
	  c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "failed to create transaction",
		"error": err,
	})
	  return
   }

   c.JSON(http.StatusOK, transaction)
}

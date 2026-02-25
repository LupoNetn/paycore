package transfer

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

   transaction, err := h.svc.CreateTransaction(c.Request.Context(), req)
   if err != nil {
	  c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
	  return
   }

   c.JSON(http.StatusOK, transaction)
}

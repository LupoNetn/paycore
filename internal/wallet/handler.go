package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luponetn/paycore/pkg/utils"
)

type Handler struct {
	Svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{Svc: svc}
}

// GetWalletHandler handles GET /wallets/:id
// It supports both wallet ID and user ID for convenience of current frontend.
func (h *Handler) GetWalletHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	authUserID := userIDVal.(uuid.UUID)

	// Try fetching as a wallet ID first
	wallet, err := h.Svc.GetWalletService(c.Request.Context(), id)
	if err == nil {
		// Security check: does the wallet belong to the user?
		if wallet.UserID != utils.ToPgUUID(authUserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this wallet"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "wallet fetched successfully", "wallet": wallet})
		return
	}

	// If not found as wallet ID, check if it's the user's own ID
	if id == authUserID {
		wallets, err := h.Svc.GetWalletsByUserService(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wallets for user"})
			return
		}
		if len(wallets) > 0 {
			// For now, return the first wallet (e.g. Savings) to satisfy current frontend
			c.JSON(http.StatusOK, gin.H{"message": "user wallets fetched", "wallet": wallets[0], "wallets": wallets})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "no wallets found for user"})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "wallet or user not found"})
}

// GetWalletTransactionsHandler handles GET /wallets/:id/transactions
func (h *Handler) GetWalletTransactionsHandler(c *gin.Context) {
	var query PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	authUserID := userIDVal.(uuid.UUID)

	var walletID uuid.UUID

	// Check if the provided ID is a wallet ID belonging to the user
	wallet, err := h.Svc.GetWalletService(c.Request.Context(), id)
	if err == nil {
		if wallet.UserID != utils.ToPgUUID(authUserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		walletID = wallet.ID
	} else if id == authUserID {
		// If it's the user ID, get their first wallet's transactions
		wallets, err := h.Svc.GetWalletsByUserService(c.Request.Context(), id)
		if err != nil || len(wallets) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no wallets found"})
			return
		}
		walletID = wallets[0].ID
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	limit := query.PageSize
	if limit <= 0 {
		limit = 10
	}
	offset := (query.Page - 1) * query.PageSize
	if offset < 0 {
		offset = 0
	}

	transactions, err := h.Svc.GetWalletTransactionsService(c.Request.Context(), walletID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "transactions fetched successfully",
		"transactions": transactions,
	})
}

func (h *Handler) GetMyWallets(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	authUserID := userIDVal.(uuid.UUID)

	wallets, err := h.Svc.GetWalletsByUserService(c.Request.Context(), authUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wallets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wallets fetched successfully", "wallets": wallets})
}

func (h *Handler) ResolveAccountHandler(c *gin.Context) {
	accountNo := c.Query("account_no")
	if accountNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account number is required"})
		return
	}

	row, err := h.Svc.ResolveAccountNumberService(c.Request.Context(), accountNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "account resolved successfully",
		"data":    row,
	})
}

package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("failed to bind json", "error", err)
		return
	}

	user, err := h.svc.SignUp(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		slog.Error("failed to create user", "error", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully, otp sent to your email",
		"data":    user,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("failed to bind json", "error", err)
		return
	}

	loginResponse, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login user"})
		slog.Error("failed to login user", "error", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user logged in successfully",
		"data":    loginResponse,
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("could not bind json for refresh request")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
	}

	newTokens, err := h.svc.Refresh(c.Request.Context(), req)
	if err != nil {
		slog.Error("unable to generate refresh token response", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Something went wrong with the server",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "new tokens generated successfully",
		"accessToken":  newTokens.AccessToken,
		"refreshToken": newTokens.RefreshToken,
	})
}

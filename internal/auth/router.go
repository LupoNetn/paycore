package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	auth := r.Group("/auth")

	auth.POST("/signup", h.SignUp)
	auth.POST("/login", h.Login)
}

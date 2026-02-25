package transfer

import "github.com/gin-gonic/gin"

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreateTransaction(c *gin.Context) {

}

package account

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) fetchUser(c *gin.Context, targetID int) {
	currentUserID := c.Keys["torn_id"].(int)

	user, err := h.service.GetAccountByTornID(c.Request.Context(), targetID, currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) GetAccountByTornID(c *gin.Context) {
	tornIDParam := c.Param("tornID")

	tornID, err := strconv.Atoi(tornIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tornID must be a whole number"})
		return
	}

	h.fetchUser(c, tornID)
}

func (h *Handler) GetCurrentAccount(c *gin.Context) {
	currentTornID := c.Keys["torn_id"].(int)

	h.fetchUser(c, currentTornID)
}

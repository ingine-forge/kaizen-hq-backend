package user

import (
	"fmt"
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

func (h *Handler) GetUserByTornID(c *gin.Context) {
	tornIDParam := c.Param("tornID")

	tornID, err := strconv.Atoi(tornIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tornID must be a whole number"})
		return
	}

	currentUserID := c.Keys["torn_id"].(int)
	fmt.Println(currentUserID)

	user, err := h.service.GetUserByTornID(c.Request.Context(), tornID, currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

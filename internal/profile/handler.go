package profile

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

func (h *Handler) FetchAndStoreProfile(c *gin.Context) {
	tornIDParam := c.Param("tornID")

	tornID, err := strconv.Atoi(tornIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	err = h.service.StoreProfileForID(c.Request.Context(), tornID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "error creating profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "profile successfully stored"})
}

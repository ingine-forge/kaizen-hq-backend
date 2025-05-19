package auth

import (
	"kaizen-hq/internal/account"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// handleRegistrationError handles various registration errors.
func (h *Handler) handleRegistrationError(c *gin.Context, err error) {
	// Mapping error messages to status codes and user-friendly error messages
	errorMap := map[string]struct {
		StatusCode int
		ErrorMsg   string
	}{
		ErrEmailAlreadyRegistered: {http.StatusConflict, ErrEmailAlreadyRegistered},
		ErrUserAlreadyRegistered:  {http.StatusConflict, ErrUserAlreadyRegistered},
		ErrInvalidAPIKey:          {http.StatusBadRequest, ErrInvalidAPIKey},
		ErrInvalidAPIKeyAccess:    {http.StatusBadRequest, ErrInvalidAPIKeyAccess},
		ErrUserNotFound:           {http.StatusBadRequest, ErrUserNotFound},
	}

	if mappedError, exists := errorMap[err.Error()]; exists {
		// Abort with the mapped status and error message
		c.AbortWithStatusJSON(mappedError.StatusCode, gin.H{"error": mappedError.ErrorMsg})
	} else {
		// Default case for unknown errors
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user := account.Account{
		Email:    req.Email,
		Password: req.Password,
		APIKey:   req.APIKey,
	}

	// Call the service to register the user
	if err := h.service.Register(c.Request.Context(), &user); err != nil {
		// Handle known errors by mapping them to appropriate responses
		h.handleRegistrationError(c, err)
		return
	}

	// Successful registration
	c.JSON(http.StatusCreated, gin.H{"status": "user created"})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	token, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// * Cookies are domain specific
	c.SetCookie("token", token, 3600, "/", h.service.config.CORS.ClientDomain, true, false)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/user-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/user-service/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context){
	var req models.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request:" + err.Error()})
		return 
	}
	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user with this email already exists" + err.Error()})
		return 
	}

	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload" + err.Error()})
		return 
	}	

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenctication failed" + err.Error()})
		return 
	}
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) GetProfile( c *gin.Context) {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticate"})
		return 
	}

	user, err := h.authService.GetUserProfile(c.Request.Context(), userEmail.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found" + err.Error()})
		return 
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request" + err.Error()})
		return 
	}

	user, err := h.authService.UpdateUserProfile(c.Request.Context(), userEmail.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update failed" + err.Error()})
		return 
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userEmail, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return 
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request" + err.Error()})
		return 
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userEmail.(string), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password change failed" + err.Error()})
		return 
	}
	c.JSON(http.StatusOK, gin.H{"message":"password changed successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request" + err.Error()})
		return 
	}

	c.JSON(http.StatusNotImplemented, nil)
}
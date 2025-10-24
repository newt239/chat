package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/auth"
)

type AuthHandler struct {
	authUseCase auth.AuthUseCase
}

func NewAuthHandler(authUseCase auth.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} auth.AuthOutput
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := auth.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	output, err := h.authUseCase.Register(c.Request.Context(), input)
	if err != nil {
		if err == auth.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, output)
}

// Login godoc
// @Summary Login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} auth.AuthOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.authUseCase.Login(c.Request.Context(), input)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to login"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} auth.AuthOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	input := auth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.authUseCase.RefreshToken(c.Request.Context(), input)
	if err != nil {
		if err == auth.ErrInvalidToken {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, output)
}

// Logout godoc
// @Summary Logout
// @Description Revoke refresh token and logout user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Logout request"
// @Success 200 {object} auth.LogoutOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	input := auth.LogoutInput{
		UserID:       userID.(string),
		RefreshToken: req.RefreshToken,
	}

	output, err := h.authUseCase.Logout(c.Request.Context(), input)
	if err != nil {
		if err == auth.ErrSessionNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, output)
}

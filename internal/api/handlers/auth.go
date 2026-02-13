package handlers

import (
	"context"
	"net/http"

	authpb "github.com/backend-app/backend/pkg/proto/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authClient authpb.AuthServiceClient
}

func NewAuthHandler(authClient authpb.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"securePassword123"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securePassword123"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string        `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type UserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"user@example.com"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя и возвращает JWT токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} AuthResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 409 {object} map[string]string "Пользователь с таким email уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Register(context.Background(), &authpb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		User: &UserResponse{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			CreatedAt: resp.User.CreatedAt,
		},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентифицирует пользователя и возвращает JWT токены
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} AuthResponse "Успешный вход"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 401 {object} map[string]string "Неверный email или пароль"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.Login(context.Background(), &authpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		User: &UserResponse{
			ID:        resp.User.Id,
			Email:     resp.User.Email,
			CreatedAt: resp.User.CreatedAt,
		},
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// Refresh godoc
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены используя валидный refresh token. Возвращает новую пару токенов.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} map[string]string "Токены успешно обновлены" example:"{\"access_token\":\"eyJhbGci...\",\"refresh_token\":\"eyJhbGci...\"}"
// @Failure 400 {object} map[string]string "Неверный формат данных" example:"{\"error\":\"invalid request format\"}"
// @Failure 401 {object} map[string]string "Невалидный или истекший refresh token" example:"{\"error\":\"invalid or expired refresh token\"}"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example:"{\"error\":\"failed to refresh token\"}"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authClient.RefreshToken(context.Background(), &authpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unauthenticated, codes.InvalidArgument:
				c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
			}
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

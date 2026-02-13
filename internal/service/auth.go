package service

import (
	"errors"

	"github.com/backend-app/backend/internal/models"
	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(email, password string) (*models.User, error) {
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (*models.User, string, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := jwt.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	if claims.Type != "refresh" {
		return "", "", errors.New("invalid token type")
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	newAccessToken, err := jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims, err := jwt.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		return uuid.Nil, err
	}

	if claims.Type != "access" {
		return uuid.Nil, errors.New("invalid token type")
	}

	return claims.UserID, nil
}

func (s *AuthService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return jwt.GenerateAccessToken(userID, s.jwtSecret)
}

func (s *AuthService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return jwt.GenerateRefreshToken(userID, s.jwtSecret)
}

package services

import (
	"context"

	"github.com/backend-app/backend/internal/models"
	"github.com/backend-app/backend/internal/repository"
	"github.com/backend-app/backend/pkg/jwt"
	authpb "github.com/backend-app/backend/pkg/proto/auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check user existence")
	}
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := user.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &authpb.RegisterResponse{
		User: &authpb.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &authpb.LoginResponse{
		User: &authpb.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	claims, err := jwt.ValidateToken(req.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	if claims.Type != "refresh" {
		return nil, status.Error(codes.InvalidArgument, "invalid token type")
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	newAccessToken, err := jwt.GenerateAccessToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := jwt.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	if claims.Type != "access" {
		return &authpb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		UserId: claims.UserID.String(),
		Valid:  true,
	}, nil
}

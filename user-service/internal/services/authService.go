package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ollatomiwa/hotelsystem/user-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/user-service/internal/repositories/postgres"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/security"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

)

type AuthService struct {
	userRepo *postgres.UserRepository
	security *security.JWTManager
}

func NewAuthService(userRepo *postgres.UserRepository, security *security.JWTManager ) *AuthService{
	return &AuthService {
		userRepo: userRepo,
		security: security,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	existingUser, _ := s.userRepo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password),bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	//creating user
	user := &models.User{
		Id: uuid.New().String(),
		Email: req.Email,
		PasswordHash: string(hashedPassword),
		FirstName: req.FirstName,
		LastName: req.LastName,
		Phone: req.Phone,
		Role: models.RoleCustomer,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmailAuth(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	//generate tokens
	accessToken, err := s.security.GenerateAccessToken(user.Id, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.security.GenerateRefreshToken(user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	user.PasswordHash = ""
	return &models.LoginResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		User: *user,
	}, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Don't return password hash
	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) UpdateUserProfile(ctx context.Context, email string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	
	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, email string, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetUserByEmailAuth(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, email, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

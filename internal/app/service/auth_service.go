package service

import (
	"context"
	"log"
	"time"

	appErr "github.com/bysoft-wallet/users/internal/app/errors"
	"github.com/bysoft-wallet/users/internal/app/jwt"
	"github.com/bysoft-wallet/users/internal/app/user"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepository user.UserRepository
	jwtService     *jwt.JWTService
}

type LoginResponse struct {
	Access  *jwt.ServiceJWT
	Refresh *jwt.ServiceJWT
}

type SignInRequest struct {
	Email    string
	Password string
}

type SignUpRequest struct {
	Email    string
	Password string
	Name     string
}

func NewAuthService(ur user.UserRepository, jwt *jwt.JWTService) *AuthService {
	return &AuthService{
		userRepository: ur,
		jwtService:     jwt,
	}
}

func (h *AuthService) SignIn(ctx context.Context, r *SignInRequest) (*LoginResponse, error) {
	userFound, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		return &LoginResponse{}, appErr.NewNotFoundError("User not found", "user-not-found")
	}

	if !user.CheckPasswordHash(r.Password, userFound.Hash) {
		return &LoginResponse{}, appErr.NewNotFoundError("User not found", "user-not-found")
	}

	return h.loginResponseForUser(userFound)
}

func (h *AuthService) SignUp(ctx context.Context, r *SignUpRequest) (*LoginResponse, error) {
	_, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		if !appErr.IsNotFound(err) {
			return &LoginResponse{}, err
		}
	} else {
		return &LoginResponse{}, appErr.NewIncorrectInputError("Email already in use", "email-already-in-use")
	}

	hash, err := user.HashPassword(r.Password)
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "create-user-error")
	}

	user := user.NewUser(
		uuid.New(),
		r.Email,
		r.Name,
		hash,
		time.Now(),
		time.Now(),
	)

	err = h.userRepository.Add(ctx, user)
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "user-saving-error")
	}

	return h.loginResponseForUser(user)
}

func (h *AuthService) loginResponseForUser(user *user.User) (*LoginResponse, error) {
	accessClaims := jwt.NewAccessClaims(
		user.UUID,
		user.Email,
		user.Name,
	)

	refreshClaims := jwt.NewRefreshClaims(
		user.UUID,
	)

	access, err := h.jwtService.CreateAccess(*accessClaims)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError("Could not authorize user", "could-not-authorize-user")
	}

	refresh, err := h.jwtService.CreateRefresh(*refreshClaims)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError("Could not authorize user", "could-not-authorize-user")
	}

	return &LoginResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}

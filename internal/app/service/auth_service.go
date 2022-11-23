package service

import (
	"context"
	"sync"
	"time"

	"github.com/bysoft-wallet/users/internal/app/currency"
	"github.com/bysoft-wallet/users/internal/app/errors"
	appErr "github.com/bysoft-wallet/users/internal/app/errors"
	"github.com/bysoft-wallet/users/internal/app/jwt"
	"github.com/bysoft-wallet/users/internal/app/user"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepository    user.UserRepository
	jwtService        *jwt.JWTService
	refreshRepository jwt.RefreshJWTRepository
	maxUserSessions   int
}

type LoginResponse struct {
	Access  *jwt.AccessJWT
	Refresh *jwt.RefreshJWT
}

type SignInRequest struct {
	Email    string
	Password string
	Ip       string
}

type SignUpRequest struct {
	Email    string
	Password string
	Name     string
	Ip       string
}

type UpdateSettingsRequest struct {
	UserUUID uuid.UUID
	Currency string
}

func NewAuthService(ur user.UserRepository, jwt *jwt.JWTService, rfr jwt.RefreshJWTRepository, mus int) *AuthService {
	return &AuthService{
		userRepository:    ur,
		jwtService:        jwt,
		refreshRepository: rfr,
		maxUserSessions:   mus,
	}
}

func (h *AuthService) SignIn(ctx context.Context, r *SignInRequest) (*LoginResponse, error) {
	userFound, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		return &LoginResponse{}, appErr.NewIncorrectInputError("User not found", "invalid-credentials")
	}

	if !user.CheckPasswordHash(r.Password, userFound.Hash) {
		return &LoginResponse{}, appErr.NewIncorrectInputError("User not found", "invalid-credentials")
	}

	return h.createTokens(ctx, userFound, r.Ip)
}

func (h *AuthService) SignUp(ctx context.Context, r *SignUpRequest) (*LoginResponse, error) {
	_, err := h.userRepository.FindByEmail(ctx, r.Email)
	if err != nil {
		if !appErr.IsNotFound(err) {
			return &LoginResponse{}, err
		}
	} else {
		return &LoginResponse{}, appErr.NewIncorrectInputError("Email already in use", "field-email-invalid")
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
		user.DefaultUserSettings(),
		time.Now(),
		time.Now(),
	)

	err = h.userRepository.Add(ctx, user)
	if err != nil {
		return &LoginResponse{}, appErr.NewAppError(err.Error(), "user-saving-error")
	}

	return h.createTokens(ctx, user, r.Ip)
}

func (h *AuthService) createTokens(ctx context.Context, user *user.User, ip string) (*LoginResponse, error) {
	accessClaims := jwt.NewAccessClaims(
		user.UUID,
		user.Email,
		user.Name,
	)

	refreshClaims := jwt.NewRefreshClaims(
		user.UUID,
	)

	var access *jwt.AccessJWT
	var refresh *jwt.RefreshJWT
	var err error
	var wg sync.WaitGroup

	go func (){
		access, err = h.jwtService.CreateAccess(*accessClaims)
		wg.Done()
	}()

	go func (){
		refresh, err = h.jwtService.CreateRefresh(*refreshClaims, ip)
		wg.Done()
	}()
	
	wg.Add(2)

	if err != nil{
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	refreshCount, err := h.refreshRepository.CountForUser(ctx, refreshClaims.UserId)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	if refreshCount > h.maxUserSessions {
		err = h.refreshRepository.DeleteForUserUUID(ctx, refreshClaims.UserId)
		if err != nil {
			return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
		}
	}

	err = h.refreshRepository.Add(ctx, refresh)
	if err != nil {
		return &LoginResponse{}, appErr.NewAuthorizationError(err.Error(), "could-not-authorize-user")
	}

	return &LoginResponse{
		Access:  access,
		Refresh: refresh,
	}, nil
}

func (h *AuthService) GetUser(ctx context.Context, user_uuid uuid.UUID) (*user.User, error) {
	return h.userRepository.FindById(ctx, user_uuid)
}

func (h *AuthService) SaveRefresh(ctx context.Context, user_uuid uuid.UUID) (*user.User, error) {
	return h.userRepository.FindById(ctx, user_uuid)
}

func (h *AuthService) Refresh(ctx context.Context, tokenString, ip string) (*LoginResponse, error) {
	refresh, err := h.jwtService.ValidateRefresh(tokenString, ip)
	if err != nil {
		return &LoginResponse{}, errors.NewAuthorizationError(err.Error(), "invalid-token")
	}

	exists, err := h.refreshRepository.Exists(ctx, refresh.Claims.UUID, refresh.Claims.UserId, ip, tokenString)
	if err != nil {
		return &LoginResponse{}, errors.NewAuthorizationError(err.Error(), "invalid-token")
	}

	if !exists {
		return &LoginResponse{}, errors.NewAuthorizationError("Refresh not found", "invalid-token")
	}

	err = h.refreshRepository.Delete(ctx, refresh.Claims.UUID)
	if err != nil {
		return &LoginResponse{}, errors.NewAuthorizationError(err.Error(), "invalid-token")
	}

	user, err := h.userRepository.FindById(ctx, refresh.Claims.UserId)
	if err != nil {
		return &LoginResponse{}, errors.NewAuthorizationError(err.Error(), "invalid-token")
	}

	return h.createTokens(ctx, user, ip)
}

func (h *AuthService) UpdateSettings(ctx context.Context, request *UpdateSettingsRequest) (*user.User, error) {
	cur, err := currency.FromString(request.Currency)
	if err != nil {
		return &user.User{}, appErr.NewIncorrectInputError("Invalid currency", "field-currency-invalid")
	}

	settings := user.NewSettings(cur)

	_, err = h.userRepository.FindById(ctx, request.UserUUID)
	if err != nil {
		return &user.User{}, err
	}

	return h.userRepository.UpdateSettings(ctx, request.UserUUID, &settings)
}

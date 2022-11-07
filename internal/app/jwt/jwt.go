package jwt

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JWTService struct {
	secret     string
	accessTTL  int
	refreshTTL int
}

type AccessClaims struct {
	UserId uuid.UUID
	Email  string
	Name   string
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserId uuid.UUID
	jwt.RegisteredClaims
}

type Claims interface{}

type ServiceJWT struct {
	Claims Claims
	Token  string
}

func NewAccessClaims(UserId uuid.UUID, Email, Name string) *AccessClaims {
	return &AccessClaims{
		UserId: UserId,
		Email:  Email,
		Name:   Name,
	}
}

func NewRefreshClaims(UserId uuid.UUID) *RefreshClaims {
	return &RefreshClaims{
		UserId: UserId,
	}
}

func NewJwtService(secret string, accessTTL, refreshTTL int) *JWTService {
	return &JWTService{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (h *JWTService) CreateAccess(c AccessClaims) (*ServiceJWT, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(h.accessTTL)))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	sign, err := token.SignedString([]byte(h.secret))
	if err != nil {
		return &ServiceJWT{}, err
	}

	return &ServiceJWT{
		Claims: c,
		Token:  sign,
	}, nil
}

func (h *JWTService) CreateRefresh(c RefreshClaims) (*ServiceJWT, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(h.refreshTTL)))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	sign, err := token.SignedString([]byte(h.secret))
	if err != nil {
		return &ServiceJWT{}, err
	}

	return &ServiceJWT{
		Claims: c,
		Token:  sign,
	}, nil
}

func (h *JWTService) ParseAccess(token string) (*ServiceJWT, error) {
	t, err := jwt.ParseWithClaims(token, &AccessClaims{}, h.validateParsed)
	if err != nil {
		return &ServiceJWT{}, nil
	}

	return &ServiceJWT{
		Claims: t.Claims,
		Token:  token,
	}, nil
}

func (h *JWTService) ParseRefresh(token string) (*ServiceJWT, error) {
	t, err := jwt.ParseWithClaims(token, &RefreshClaims{}, h.validateParsed)
	if err != nil {
		return &ServiceJWT{}, nil
	}

	return &ServiceJWT{
		Claims: t.Claims,
		Token:  token,
	}, nil
}

func (h *JWTService) validateParsed(parsed *jwt.Token) (interface{}, error) {
	if _, ok := parsed.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", parsed.Header["alg"])
	}

	return []byte(h.secret), nil
}

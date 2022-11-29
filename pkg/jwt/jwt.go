package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

type JWTService struct {
	secret     string
	accessTTL  int
	refreshTTL int
}

type AccessClaims struct {
	UserId uuid.UUID
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UUID   uuid.UUID
	UserId uuid.UUID
	jwt.RegisteredClaims
}

type AccessJWT struct {
	Claims AccessClaims
	Token  string
}

type RefreshJWT struct {
	Claims RefreshClaims
	Token  string
	Ip     string
}

type RefreshJWTRepository interface {
	Add(ctx context.Context, refresh *RefreshJWT) error
	Exists(ctx context.Context, uuid, userUUID uuid.UUID, ip string, token string) (bool, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	DeleteForUserUUID(ctx context.Context, userUUID uuid.UUID) error
	CountForUser(ctx context.Context, userUUID uuid.UUID) (int, error)
}

func NewAccessClaims(UserId uuid.UUID, Email, Name string) *AccessClaims {
	return &AccessClaims{
		UserId: UserId,
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

func (h *JWTService) CreateAccess(c AccessClaims) (*AccessJWT, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(h.accessTTL) * time.Second))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	sign, err := token.SignedString([]byte(h.secret))
	if err != nil {
		return &AccessJWT{}, err
	}

	return &AccessJWT{
		Claims: c,
		Token:  sign,
	}, nil
}

func (h *JWTService) CreateRefresh(c RefreshClaims, ip string) (*RefreshJWT, error) {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(h.refreshTTL) * time.Second))
	c.UUID = uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	sign, err := token.SignedString([]byte(h.secret))
	if err != nil {
		return &RefreshJWT{}, err
	}

	return &RefreshJWT{
		Claims: c,
		Token:  sign,
		Ip:     ip,
	}, nil
}

func (h *JWTService) ValidateAccess(token string) (*AccessJWT, error) {
	t, err := jwt.ParseWithClaims(token, &AccessClaims{}, h.validateParsed)
	if err != nil {
		return &AccessJWT{}, err
	}

	if !t.Valid {
		return &AccessJWT{}, errors.New("invalid token")
	}

	return &AccessJWT{
		Claims: *t.Claims.(*AccessClaims),
		Token:  token,
	}, nil
}

func (h *JWTService) ValidateRefresh(token, ip string) (*RefreshJWT, error) {
	t, err := jwt.ParseWithClaims(token, &RefreshClaims{}, h.validateParsed)
	if err != nil {
		return &RefreshJWT{}, nil
	}
	claims := *t.Claims.(*RefreshClaims)

	if !t.Valid {
		return &RefreshJWT{}, errors.New("invalid token")
	}

	return &RefreshJWT{
		Claims: claims,
		Token:  token,
		Ip:     ip,
	}, nil
}

func (h *JWTService) validateParsed(parsed *jwt.Token) (interface{}, error) {
	if _, ok := parsed.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", parsed.Header["alg"])
	}

	return []byte(h.secret), nil
}

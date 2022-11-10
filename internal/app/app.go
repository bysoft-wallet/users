package app

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/bysoft-wallet/users/internal/adapters"
	"github.com/bysoft-wallet/users/internal/app/jwt"
	"github.com/bysoft-wallet/users/internal/app/service"
	"github.com/jackc/pgx/v5"
)

type Application struct {
	AuthService *service.AuthService
	JWTService  *jwt.JWTService
}

func NewApplication(ctx context.Context) (*Application, error) {
	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		return &Application{}, err
	}

	JWTSecret := os.Getenv("JWT_SECRET")
	JWTAccessTTL, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	JWTRefreshTTL, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if JWTSecret == "" || err != nil {
		return &Application{}, errors.New("JWT configuration must be provided")
	}

	jwtService := jwt.NewJwtService(
		JWTSecret,
		JWTAccessTTL,
		JWTRefreshTTL,
		adapters.NewRefreshPgsqlRepository(conn),
	)

	authService := service.NewAuthService(
		adapters.NewUserPgsqlRepository(conn),
		jwtService,
	)

	return &Application{
		AuthService: authService,
		JWTService:  jwtService,
	}, nil
}

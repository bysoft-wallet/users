package app

import (
	"context"
	"github.com/bysoft-wallet/users/internal/adapters"
	"github.com/bysoft-wallet/users/internal/app/service"
	"github.com/bysoft-wallet/users/pkg/jwt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Application struct {
	AuthService *service.AuthService
	JWTService  *jwt.JWTService
	Logger      *logrus.Logger
}

type Config struct {
	Ctx             context.Context
	Logger          *logrus.Logger
	DbPool          *pgxpool.Pool
	JwtSecret       string
	JwtAccessTTL    *int
	JwtRefreshTTL   *int
	MaxUserSessions int
}

func NewApplication(config *Config) (*Application, error) {

	jwtService := jwt.NewJwtService(&jwt.JWTConfig{
			Secret: config.JwtSecret,
			AccessTTL: config.JwtAccessTTL,
			RefreshTTL: config.JwtRefreshTTL,
		},
	)

	authService := service.NewAuthService(
		adapters.NewUserPgsqlRepository(config.DbPool),
		jwtService,
		adapters.NewRefreshPgsqlRepository(config.DbPool),
		config.MaxUserSessions,
	)

	return &Application{
		AuthService: authService,
		JWTService:  jwtService,
		Logger:      config.Logger,
	}, nil
}

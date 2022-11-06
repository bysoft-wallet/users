package app

import (
	"context"
	"os"

	"github.com/bysoft-wallet/users/internal/adapters"
	"github.com/bysoft-wallet/users/internal/app/service"
	"github.com/jackc/pgx/v5"
)

type Application struct {
	UserService *service.UserService
}

func NewApplication(ctx context.Context) (*Application, error) {
	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		return &Application{}, err
	}

	return &Application{
		UserService: service.NewUserService(
			adapters.NewUserPgsqlRepository(conn),
		),
	}, nil
}

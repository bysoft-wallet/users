package adapters

import (
	"context"
	"time"

	"github.com/bysoft-wallet/users/internal/app/service"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserModel struct {
	UUID      uuid.UUID `db:uuid`
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserPgsqlRepository struct {
	conn *pgx.Conn
}

func NewUserPgsqlRepository(conn *pgx.Conn) *UserPgsqlRepository {
	return &UserPgsqlRepository{conn}
}

func (s *UserPgsqlRepository) FindById(ctx context.Context, uuid uuid.UUID) (*service.User, error) {
	var userModel UserModel
	if err := pgxscan.Get(
		ctx, s.conn, userModel, "select * from users where id = $1", uuid,
	); err != nil {
		return &service.User{}, nil
	}

	return serviceUserFromModel(&userModel)
}

func serviceUserFromModel(model *UserModel) (*service.User, error) {
	return service.NewUser(
		model.UUID,
		model.Email,
		model.Name,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

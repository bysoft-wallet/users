package adapters

import (
	"context"
	"time"

	"github.com/bysoft-wallet/users/internal/app/errors"
	"github.com/bysoft-wallet/users/internal/app/user"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserModel struct {
	UUID      uuid.UUID `db:"uuid"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Hash      string    `db:"hash"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserPgsqlRepository struct {
	conn *pgx.Conn
}

func NewUserPgsqlRepository(conn *pgx.Conn) *UserPgsqlRepository {
	return &UserPgsqlRepository{conn}
}

func (s *UserPgsqlRepository) FindById(ctx context.Context, uuid uuid.UUID) (*user.User, error) {
	userModel := &UserModel{}
	if err := pgxscan.Get(
		ctx, s.conn, userModel, "select * from users where id = $1", uuid,
	); err != nil {
		if pgxscan.NotFound(err) {
			return &user.User{}, errors.NewNotFoundError("User not found", "user-not-found")
		}

		return &user.User{}, err
	}

	return serviceUserFromModel(userModel)
}

func (s *UserPgsqlRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	userModel := &UserModel{}
	if err := pgxscan.Get(
		ctx, s.conn, userModel, "select * from users where email = $1", email,
	); err != nil {
		if pgxscan.NotFound(err) {
			return &user.User{}, errors.NewNotFoundError("User not found", "user-not-found")
		}

		return &user.User{}, err
	}

	return serviceUserFromModel(userModel)
}

func (s *UserPgsqlRepository) Add(ctx context.Context, u *user.User) error {
	_, err := s.conn.Exec(ctx, "insert into users(uuid, email, name, hash, created_at, updated_at) values($1,$2,$3,$4,$5,$6)", u.UUID, u.Email, u.Name, u.Hash, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func serviceUserFromModel(model *UserModel) (*user.User, error) {
	return user.NewUser(
		model.UUID,
		model.Email,
		model.Name,
		model.Hash,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

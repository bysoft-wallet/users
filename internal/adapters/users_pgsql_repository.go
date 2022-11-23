package adapters

import (
	"context"
	"time"

	"github.com/bysoft-wallet/users/internal/app/currency"
	"github.com/bysoft-wallet/users/internal/app/errors"
	"github.com/bysoft-wallet/users/internal/app/user"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserModel struct {
	UUID      uuid.UUID         `db:"uuid"`
	Email     string            `db:"email"`
	Name      string            `db:"name"`
	Hash      string            `db:"hash"`
	Settings  map[string]string `db:"settings"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type UserPgsqlRepository struct {
	pool *pgxpool.Pool
}

func UserSettingsToMap(s user.Settings) map[string]string {
	return map[string]string{
		"currency": s.Currency.String(),
	}
}

func NewUserPgsqlRepository(pool *pgxpool.Pool) *UserPgsqlRepository {
	return &UserPgsqlRepository{pool}
}

func (s *UserPgsqlRepository) FindById(ctx context.Context, uuid uuid.UUID) (*user.User, error) {
	userModel := &UserModel{}
	if err := pgxscan.Get(
		ctx, s.pool, userModel, "select uuid, email, name, hash, settings, created_at, updated_at from users where uuid = $1", uuid,
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
		ctx, s.pool, userModel, "select uuid, email, name, hash, settings, created_at, updated_at from users where email = $1", email,
	); err != nil {
		if pgxscan.NotFound(err) {
			return &user.User{}, errors.NewNotFoundError("User not found", "user-not-found")
		}

		return &user.User{}, err
	}

	return serviceUserFromModel(userModel)
}

func (s *UserPgsqlRepository) Add(ctx context.Context, u *user.User) error {
	_, err := s.pool.Exec(ctx, "insert into users(uuid, email, name, hash, settings, created_at, updated_at) values($1,$2,$3,$4,$5,$6, $7)", u.UUID, u.Email, u.Name, u.Hash, UserSettingsToMap(u.Settings), u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserPgsqlRepository) UpdateSettings(ctx context.Context, user_uuid uuid.UUID, settings *user.Settings) (*user.User, error) {
	_, err := s.pool.Exec(ctx, "update users set settings = $1 where uuid = $2", UserSettingsToMap(*settings), user_uuid)
	if err != nil {
		return &user.User{}, err
	}

	return s.FindById(ctx, user_uuid)
}

func serviceUserFromModel(model *UserModel) (*user.User, error) {
	cur, err := currency.FromString(model.Settings["currency"])
	if err != nil {
		return &user.User{}, err
	}

	return user.NewUser(
		model.UUID,
		model.Email,
		model.Name,
		model.Hash,
		user.NewSettings(cur),
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

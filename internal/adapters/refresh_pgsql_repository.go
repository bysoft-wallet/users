package adapters

import (
	"context"
	"time"

	"github.com/bysoft-wallet/users/internal/app/jwt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RefreshModel struct {
	UUID      uuid.UUID `db:"uuid"`
	UserUUID  uuid.UUID `db:"user_uuid"`
	Token     string    `db:"token"`
	Ip        string    `db:"ip"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RefreshPgsqlRepository struct {
	conn *pgx.Conn
}

func NewRefreshPgsqlRepository(conn *pgx.Conn) *RefreshPgsqlRepository {
	return &RefreshPgsqlRepository{conn}
}

func (s *RefreshPgsqlRepository) Add(ctx context.Context, refresh *jwt.RefreshJWT) error {
	_, err := s.conn.Exec(ctx, "insert into refresh_tokens(uuid, user_uuid, token, ip, created_at, updated_at) values($1,$2,$3,$4,$5,$6)",
		refresh.Claims.UUID,
		refresh.Claims.UserId,
		refresh.Token,
		refresh.Ip,
		time.Now(),
		time.Now())

	if err != nil {
		return err
	}

	return nil
}
func (s *RefreshPgsqlRepository) Exists(ctx context.Context, uuid, userUUID uuid.UUID, ip string, token string) (bool, error) {
	model := &RefreshModel{}
	if err := pgxscan.Get(
		ctx, s.conn, model, "select * from refresh_tokens where uuid = $1 and user_uuid = $2 and ip = $3 and token = $4",
		uuid,
		userUUID,
		ip,
		token,
	); err != nil {
		if pgxscan.NotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *RefreshPgsqlRepository) Delete(ctx context.Context, uuid uuid.UUID) error {
	_, err := s.conn.Exec(ctx, "delete from refresh_tokens where uuid = $1", uuid)

	if err != nil {
		return err
	}

	return nil
}

func (s *RefreshPgsqlRepository) DeleteForUserUUID(ctx context.Context, userUUID uuid.UUID) error {
	_, err := s.conn.Exec(ctx, "delete from refresh_tokens where user_uuid = $1", userUUID)

	if err != nil {
		return err
	}

	return nil
}

func (s *RefreshPgsqlRepository) CountForUser(ctx context.Context, userUUID uuid.UUID) (int, error) {
	var counter int

	err := s.conn.QueryRow(ctx, "SELECT count(*) FROM refresh_tokens where user_uuid = $1", userUUID).Scan(&counter)
	return counter, err
}

func serviceRefreshJWTFromModel(model *RefreshModel) (*jwt.RefreshJWT, error) {
	claims := &jwt.RefreshClaims{
		UUID:   model.UUID,
		UserId: model.UserUUID,
	}

	return &jwt.RefreshJWT{
		Claims: *claims,
		Token:  model.Token,
		Ip:     model.Ip,
	}, nil
}

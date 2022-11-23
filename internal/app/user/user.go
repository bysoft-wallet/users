package user

import (
	"context"
	"time"

	"github.com/bysoft-wallet/users/internal/app/currency"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID      uuid.UUID
	Email     string
	Name      string
	Hash      string
	Settings  Settings
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Settings struct {
	Currency currency.Currency
}

func DefaultUserSettings() Settings {
	return Settings{
		Currency: currency.RUB,
	}
}

func NewSettings(cur currency.Currency) Settings {
	return Settings{
		Currency: cur,
	}
}

func NewUser(
	UUID uuid.UUID,
	Email string,
	Name string,
	Hash string,
	Settings Settings,
	CreatedAt time.Time,
	UpdatedAt time.Time,
) *User {
	return &User{
		UUID:      UUID,
		Email:     Email,
		Name:      Name,
		Hash:      Hash,
		Settings:  Settings,
		CreatedAt: CreatedAt,
		UpdatedAt: UpdatedAt,
	}
}

type UserService struct {
	UserRepository UserRepository
}

type UserRepository interface {
	FindById(ctx context.Context, uuid uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Add(ctx context.Context, user *User) error
	UpdateSettings(ctx context.Context, user_uuid uuid.UUID, settings *Settings) (*User, error)
}

func NewUserService(uRepo UserRepository) *UserService {
	return &UserService{UserRepository: uRepo}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

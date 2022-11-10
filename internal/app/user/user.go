package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID      uuid.UUID
	Email     string
	Name      string
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	UUID uuid.UUID,
	Email string,
	Name string,
	Hash string,
	CreatedAt time.Time,
	UpdatedAt time.Time,
) *User {
	return &User{
		UUID:      UUID,
		Email:     Email,
		Name:      Name,
		Hash:	   Hash,
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
}

func NewUserService(uRepo UserRepository) *UserService {
	return &UserService{UserRepository: uRepo}
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

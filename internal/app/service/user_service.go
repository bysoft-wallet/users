package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	UUID uuid.UUID,
	Email string,
	Name string,
	CreatedAt time.Time,
	UpdatedAt time.Time,
) *User {
	return &User{
		UUID:      UUID,
		Email:     Email,
		Name:      Name,
		CreatedAt: CreatedAt,
		UpdatedAt: UpdatedAt,
	}
}

type UserService struct {
	UserRepository UserRepository
}

type UserRepository interface {
	FindById(ctx context.Context, uuid uuid.UUID) (*User, error)
}

func NewUserService(uRepo UserRepository) *UserService {
	return &UserService{UserRepository: uRepo}
}

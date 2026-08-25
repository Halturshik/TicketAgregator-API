package users

import (
	"context"
	"time"
)

type Service interface {
	GetProfile(ctx context.Context, userID int) (*Profile, error)
	UpdateProfile(ctx context.Context, userID int, in UpdateProfileInput) (*Profile, error)
}

type Repository interface {
	GetProfile(ctx context.Context, userID int) (*Profile, error)
	UpdateProfile(ctx context.Context, userID int, in UpdateProfileInput, birthDate time.Time) (*Profile, error)
}

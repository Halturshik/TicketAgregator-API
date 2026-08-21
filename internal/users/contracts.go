package users

import "context"

type Service interface {
	GetProfile(ctx context.Context, userID int) (*Profile, error)
	UpdateProfile(ctx context.Context, userID int, in UpdateProfileInput) (*Profile, error)
}

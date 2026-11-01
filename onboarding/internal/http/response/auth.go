package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
)

type User struct {
	Id        uuid.UUID `json:"id" format:"uuid"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type Auth struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func NewUser(
	user *domains.User,
) *User {
	if user == nil {
		return nil
	}

	return &User{
		Id:        user.Id,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func NewAuth(
	user *domains.User,
	token string,
) *Auth {
	return &Auth{
		Token: token,
		User:  NewUser(user),
	}
}

package auth

import (
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email       string                  `json:"email"       validate:"required,email,max=254"`
	Password    string                  `json:"password"    validate:"required,min=8,max=72"`
	DisplayName httpx.Optional[string] `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email,max=254"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Role        string    `json:"role"`
	Email       *string   `json:"email"`
	DisplayName *string   `json:"displayName"`
	CreatedAt   string    `json:"createdAt"`
	ConvertedAt *string   `json:"convertedAt"`
}

func ToUserResponse(u db.User) UserResponse {
	resp := UserResponse{
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	resp.ID, _ = uuid.FromBytes(u.ID.Bytes[:])
	if u.Email.Valid {
		v := u.Email.String
		resp.Email = &v
	}
	if u.DisplayName.Valid {
		v := u.DisplayName.String
		resp.DisplayName = &v
	}
	if u.ConvertedAt.Valid {
		v := u.ConvertedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		resp.ConvertedAt = &v
	}
	return resp
}

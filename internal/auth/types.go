package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SignUpRequest struct {
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Nationality string `json:"nationality"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          uuid.UUID          `json:"id"`
	FullName    string             `json:"full_name"`
	PhoneNumber string             `json:"phone_number"`
	Email       string             `json:"email"`
	Username    string             `json:"username"`
	AccountNo   string             `json:"account_no"`
	Nationality string             `json:"nationality"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type CreateOTPRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type OTPResponse struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	Code      string             `json:"code"`
	Purpose   string             `json:"purpose"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	Used      pgtype.Bool        `json:"used"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

package auth

import (
	"context"
	"log/slog"

	"github.com/luponetn/paycore/internal/config"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/pkg/utils"
)

type Svc struct {
	queries *db.Queries
	cfg     *config.Config
}

type Service interface {
	SignUp(ctx context.Context, req SignUpRequest) (UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
}

func NewService(queries *db.Queries, cfg *config.Config) Service {
	return &Svc{queries: queries, cfg: cfg}
}

// SignUp handles the business logic for user registration
func (s *Svc) SignUp(ctx context.Context, req SignUpRequest) (UserResponse, error) {
	//  Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return UserResponse{}, err
	}

	// Generate unique username
	username := utils.GenerateUsername(req.FullName)

	// Generate account number from phone number
	accountNo := utils.GenerateAccountNumber(req.PhoneNumber)

	// Create the user in the database
	arg := db.CreateUserParams{
		FullName:     req.FullName,
		PhoneNumber:  req.PhoneNumber,
		Email:        req.Email,
		Passwordhash: hashedPassword,
		Username:     username,
		AccountNo:    accountNo,
		Nationality:  req.Nationality,
	}

	user, err := s.queries.CreateUser(ctx, arg)
	if err != nil {
		return UserResponse{}, err
	}

	return UserResponse{
		ID:          user.ID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
		Username:    user.Username,
		AccountNo:   user.AccountNo,
		Nationality: user.Nationality,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// login handles the business logic for user login
func (s *Svc) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		slog.Error("failed to get user by email", "error", err)
		return LoginResponse{}, err
	}

	if err := utils.CheckPassword(req.Password, user.Passwordhash); err != nil {
		slog.Error("failed to check password", "error", err)
		return LoginResponse{}, err
	}

	accessToken, err := utils.GenerateToken(user.ID, user.Username, s.cfg.JWTAccessSecret, "access")
	if err != nil {
		return LoginResponse{}, err
	}

	refreshToken, err := utils.GenerateToken(user.ID, user.Username, s.cfg.JWTRefreshSecret, "refresh")
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		User: UserResponse{
			ID:          user.ID,
			FullName:    user.FullName,
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
			Username:    user.Username,
			AccountNo:   user.AccountNo,
			Nationality: user.Nationality,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

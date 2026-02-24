package auth

import (
	"context"
	"log/slog"

	// "time"

	// "github.com/jackc/pgx/v5/pgtype"

	// "github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luponetn/paycore/internal/config"
	"github.com/luponetn/paycore/internal/db"
	"github.com/luponetn/paycore/pkg/utils"
)

type Svc struct {
	db         *pgxpool.Pool
	queries    *db.Queries
	cfg        *config.Config
	taskClient *asynq.Client
}

type Service interface {
	SignUp(ctx context.Context, req SignUpRequest) (UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (RefreshResponse, error)
	//CreateOTP(ctx context.Context, userID uuid.UUID) (OTPResponse, error)
}

func NewService(queries *db.Queries, cfg *config.Config, taskClient *asynq.Client, db *pgxpool.Pool) Service {
	return &Svc{db: db, queries: queries, cfg: cfg, taskClient: taskClient}
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

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return UserResponse{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, arg)
	if err != nil {
		return UserResponse{}, err
	}

	//TODO: create wallet for the user
	walletTypes := []db.WalletTypeEnum{
		db.WalletTypeEnumSavings,
		db.WalletTypeEnumFixed,
		db.WalletTypeEnumMisc,
	}

	for _, value := range walletTypes {
		_, err := qtx.CreateWallet(ctx, db.CreateWalletParams{
			UserID:     utils.ToPgUUID(user.ID),
			WalletType: db.WalletTypeEnum(value),
			Currency:   "NGN",
		})

		if err != nil {
			slog.Error("could not create wallets for user", "error", err)
			return UserResponse{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
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

// refres token endpoint
func (s *Svc) Refresh(ctx context.Context, req RefreshRequest) (RefreshResponse, error) {
	claims, err := utils.VerifyToken(req.RefreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		slog.Error("could not verify refresh token for user", "error", err)
		return RefreshResponse{}, err
	}

	newAccessToken, accessErr := utils.GenerateToken(claims.UserID, claims.Username, s.cfg.JWTAccessSecret, "access")
	if accessErr != nil {
		slog.Error("could not generate new access token for refresh service", "error", accessErr)
		return RefreshResponse{}, accessErr
	}
	newRefreshToken, refreshErr := utils.GenerateToken(claims.UserID, claims.Username, s.cfg.JWTRefreshSecret, "refresh")
	if refreshErr != nil {
		slog.Error("could not generate new refresh token for refresh service", "error", refreshErr)
		return RefreshResponse{}, accessErr
	}

	return RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// createOtp handles the creation of otp in the database
// func (s *Svc) CreateOTP(ctx context.Context, userID uuid.UUID) (OTPResponse, error) {
// 	otp, err := utils.GenerateOTP()
// 	if err != nil {
// 		slog.Error("failed to generate otp", "error", err)
// 		return OTPResponse{}, err
// 	}

// 	expiresAt := time.Now().Add(time.Minute * 10)
// 	payload := db.CreateOTPParams{
// 		UserID:  userID,
// 		Code:    otp,
// 		Purpose: "verification",
// 		ExpiresAt: pgtype.Timestamptz{
// 			Time:  expiresAt,
// 			Valid: true,
// 		},
// 		Used: pgtype.Bool{
// 			Bool:  false,
// 			Valid: true,
// 		},
// 	}

// 	otpRecord, err := s.queries.CreateOTP(ctx, payload)
// 	if err != nil {
// 		slog.Error("failed to create otp", "error", err)
// 		return OTPResponse{}, err
// 	}

// 	return OTPResponse{
// 		ID:        otpRecord.ID,
// 		UserID:    otpRecord.UserID,
// 		Code:      otpRecord.Code,
// 		Purpose:   otpRecord.Purpose,
// 		ExpiresAt: otpRecord.ExpiresAt,
// 		Used:      otpRecord.Used,
// 	}, nil

// }

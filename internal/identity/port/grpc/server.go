package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/alexdolgov/auth-service/internal/identity/application/command"
	"github.com/alexdolgov/auth-service/internal/identity/domain"
	pb "github.com/alexdolgov/auth-service/internal/identity/port/grpc/pb/auth/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer

	register     *command.RegisterHandler
	login        *command.LoginHandler
	logout       *command.LogoutHandler
	requestReset *command.RequestResetHandler
	confirmReset *command.ConfirmResetHandler
	refreshToken *command.RefreshTokenHandler
	loginLimiter domain.RateLimiter
	resetLimiter domain.RateLimiter
}

func NewAuthServer(
	register *command.RegisterHandler,
	login *command.LoginHandler,
	logout *command.LogoutHandler,
	requestReset *command.RequestResetHandler,
	confirmReset *command.ConfirmResetHandler,
	refreshToken *command.RefreshTokenHandler,
	loginLimiter domain.RateLimiter,
	resetLimiter domain.RateLimiter,
) *AuthServer {
	return &AuthServer{
		register:     register,
		login:        login,
		logout:       logout,
		requestReset: requestReset,
		confirmReset: confirmReset,
		refreshToken: refreshToken,
		loginLimiter: loginLimiter,
		resetLimiter: resetLimiter,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.register.Handle(ctx, command.RegisterUser{
		Email:           req.Email,
		Password:        req.Password,
		PasswordConfirm: req.PasswordConfirm,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.RegisterResponse{
		UserId: user.ID.String(),
		Email:  user.Email.String(),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if !s.loginLimiter.Allow(req.Email) {
		return nil, status.Error(codes.ResourceExhausted, domain.ErrRateLimitExceeded.Error())
	}

	result, err := s.login.Handle(ctx, command.LoginUser{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	s.loginLimiter.Reset(req.Email)

	return &pb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionId:    result.SessionID.String(),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id format")
	}
	if err := s.logout.Handle(ctx, command.LogoutUser{SessionID: sessionID}); err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.LogoutResponse{}, nil
}

func (s *AuthServer) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.RequestPasswordResetResponse, error) {
	if !s.resetLimiter.Allow(req.Email) {
		return nil, status.Error(codes.ResourceExhausted, domain.ErrRateLimitExceeded.Error())
	}

	rawToken, err := s.requestReset.Handle(ctx, command.RequestPasswordReset{Email: req.Email})
	if err != nil {
		return nil, mapDomainError(err)
	}

	slog.Info("mock: password reset token (would be sent via email in production)",
		"email", req.Email, "raw_token", rawToken)

	return &pb.RequestPasswordResetResponse{
		Message: "If the email exists, a reset link has been sent. Check server logs for mock token.",
	}, nil
}

func (s *AuthServer) ConfirmPasswordReset(ctx context.Context, req *pb.ConfirmPasswordResetRequest) (*pb.ConfirmPasswordResetResponse, error) {
	err := s.confirmReset.Handle(ctx, command.ConfirmPasswordReset{
		Token:           req.Token,
		NewPassword:     req.NewPassword,
		ConfirmPassword: req.ConfirmPassword,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.ConfirmPasswordResetResponse{
		Message: "Password has been successfully reset.",
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	result, err := s.refreshToken.Handle(ctx, command.RefreshToken{Token: req.RefreshToken})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		SessionId:    result.SessionID.String(),
	}, nil
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidEmailFormat):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPasswordTooShort):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPasswordTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPasswordMismatch):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, domain.ErrSessionNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, domain.ErrSessionExpired):
		return status.Error(codes.Unauthenticated, "session expired")
	case errors.Is(err, domain.ErrResetTokenNotFound):
		return status.Error(codes.InvalidArgument, "invalid or expired token")
	case errors.Is(err, domain.ErrResetTokenExpired):
		return status.Error(codes.InvalidArgument, "invalid or expired token")
	case errors.Is(err, domain.ErrResetTokenAlreadyUsed):
		return status.Error(codes.InvalidArgument, "invalid or expired token")
	case errors.Is(err, domain.ErrResetCooldown):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, domain.ErrRateLimitExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

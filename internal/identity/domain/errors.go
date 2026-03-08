package domain

import "errors"

var (
	ErrEmailAlreadyTaken     = errors.New("email address is already taken")
	ErrInvalidEmailFormat    = errors.New("email address has invalid format")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong       = errors.New("password must be at most 72 characters")
	ErrPasswordMismatch      = errors.New("passwords do not match")
	ErrInvalidCredentials    = errors.New("invalid email or password combination")
	ErrUserNotFound          = errors.New("user not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionExpired        = errors.New("session has expired")
	ErrResetTokenNotFound    = errors.New("reset token not found")
	ErrResetTokenExpired     = errors.New("reset token has expired")
	ErrResetTokenAlreadyUsed = errors.New("reset token has already been used")
	ErrResetCooldown         = errors.New("password reset was requested too recently, please wait")
	ErrRateLimitExceeded     = errors.New("too many attempts, please try again later")
)

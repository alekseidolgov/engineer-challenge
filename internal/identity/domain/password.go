package domain

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72 // bcrypt/argon2 практический лимит; защита от DoS через длинный input
)

type Password struct {
	raw string
}

func NewPassword(raw string) (Password, error) {
	if len(raw) < MinPasswordLength {
		return Password{}, ErrPasswordTooShort
	}
	if len(raw) > MaxPasswordLength {
		return Password{}, ErrPasswordTooLong
	}
	return Password{raw: raw}, nil
}

func NewPasswordPair(password, confirm string) (Password, error) {
	if password != confirm {
		return Password{}, ErrPasswordMismatch
	}
	return NewPassword(password)
}

func (p Password) Raw() string { return p.raw }

package domain

import (
	"net/mail"
	"strings"
)

type Email struct {
	address string
}

const MaxEmailLength = 254 // RFC 5321: максимальная длина email-адреса

func NewEmail(raw string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" || len(normalized) > MaxEmailLength {
		return Email{}, ErrInvalidEmailFormat
	}
	addr, err := mail.ParseAddress(normalized)
	if err != nil || addr.Address != normalized {
		return Email{}, ErrInvalidEmailFormat
	}
	// mail.ParseAddress пропускает адреса вида "user@localhost" —
	// для реального auth-сервиса требуем точку в домене.
	if !strings.Contains(strings.SplitN(normalized, "@", 2)[1], ".") {
		return Email{}, ErrInvalidEmailFormat
	}
	return Email{address: normalized}, nil
}

func (e Email) String() string { return e.address }

func (e Email) Equals(other Email) bool {
	return e.address == other.address
}

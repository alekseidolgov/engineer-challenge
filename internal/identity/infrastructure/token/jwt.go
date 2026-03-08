package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/alexdolgov/auth-service/internal/identity/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTIssuer struct {
	privateKey *ecdsa.PrivateKey
	ttl        time.Duration
}

// NewJWTIssuer создаёт issuer с существующим ключом (например, из файла/vault).
func NewJWTIssuer(privateKey *ecdsa.PrivateKey, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{privateKey: privateKey, ttl: ttl}
}

// GenerateKey создаёт новую пару ключей ES256 (P-256).
// В production ключ должен загружаться из secure storage, а не генерироваться при старте.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func (j *JWTIssuer) Issue(userID domain.UserID, email string) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(j.ttl).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return t.SignedString(j.privateKey)
}

func (j *JWTIssuer) Validate(tokenStr string) (*domain.AccessTokenClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return &j.privateKey.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	sub, _ := claims.GetSubject()
	uid, err := uuid.Parse(sub)
	if err != nil {
		return nil, err
	}
	email, _ := claims["email"].(string)
	return &domain.AccessTokenClaims{UserID: uid, Email: email}, nil
}

// PublicKey возвращает публичный ключ для верификации другими сервисами.
func (j *JWTIssuer) PublicKey() *ecdsa.PublicKey {
	return &j.privateKey.PublicKey
}

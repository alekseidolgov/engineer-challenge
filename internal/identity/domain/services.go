package domain

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

type AccessTokenIssuer interface {
	Issue(userID UserID, email string) (string, error)
	Validate(token string) (*AccessTokenClaims, error)
}

type AccessTokenClaims struct {
	UserID UserID
	Email  string
}

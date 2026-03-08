package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MiB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16

	// PHC format: $argon2id$v=X$m=M,t=T,p=P$salt$hash
	argon2PHCPartsCount = 6
	argon2PHCVersionIdx = 2
	argon2PHCParamsIdx  = 3
	argon2PHCSaltIdx    = 4
	argon2PHCHashIdx    = 5
)

type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen uint32
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    argon2Time,
		memory:  argon2Memory,
		threads: argon2Threads,
		keyLen:  argon2KeyLen,
		saltLen: argon2SaltLen,
	}
}

func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)
	// Формат PHC: параметры в хеше — не секрет, а необходимость.
	// Позволяет верифицировать старые хеши после смены параметров (миграция без сброса паролей).
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.time, h.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Argon2Hasher) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != argon2PHCPartsCount {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[argon2PHCVersionIdx], "v=%d", &version); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(parts[argon2PHCParamsIdx], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[argon2PHCSaltIdx])
	if err != nil {
		return false, err
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[argon2PHCHashIdx])
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(storedHash)))
	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1, nil
}

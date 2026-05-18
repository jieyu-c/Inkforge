package password

import (
	"golang.org/x/crypto/bcrypt"
)

func Hash(secret string, cost int) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(secret), cost)
	return string(b), err
}

func Verify(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

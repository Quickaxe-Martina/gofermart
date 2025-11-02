package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// GenerateHash generate password hash
func GenerateHash(password string) (string, error) {
	cost := bcrypt.DefaultCost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassowrd verify password hash
func VerifyPassowrd(password string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

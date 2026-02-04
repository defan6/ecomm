package util

import "golang.org/x/crypto/bcrypt"

type BcryptPasswordEncoder struct{}

// NewPasswordEncoder создаёт новый экземпляр BcryptPasswordEncoder
func NewPasswordEncoder() *BcryptPasswordEncoder {
	return &BcryptPasswordEncoder{}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
	return err == nil
}

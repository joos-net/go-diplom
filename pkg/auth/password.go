package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword создает хеш пароля используя bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// VerifyPassword проверяет что пароль соответствует хешу
func VerifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

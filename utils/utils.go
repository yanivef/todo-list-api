package utils

import (
	"database/sql"
	"os"
	"regexp"
	"task-manager-api/models"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func IsValidEmail(email string) bool {
	// define regex for valid email
	const emailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailPattern)
	return re.MatchString(email)
}

func IsEmailExists(email string, DB *sql.DB) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`
	var exists bool
	err := DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GenerateToken(user models.Users) (string, error) {
	claims := &jwt.StandardClaims{
		// subject define who the token is for
		Subject: user.Email,
		// expiresAt sets the expiration time for the token (24 hours from the current time in this case)
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}

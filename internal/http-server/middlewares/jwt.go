package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"todo_restapi/internal/config"
)

type AuthService struct {
	Config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{Config: cfg}
}

func (a *AuthService) validatePWD(password string) (bool, error) {

	if password != a.Config.Password {
		return false, errors.New("invalid password")
	}

	return true, nil
}

func (a *AuthService) GenerateJWT(password string) (string, error) {

	if ok, err := a.validatePWD(password); !ok {
		return "", fmt.Errorf("ValidatePWD: function error: %w", err)
	}

	hash := sha256.New()
	hash.Write([]byte(password))
	hashPassword := hex.EncodeToString(hash.Sum(nil))

	payload := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 8).Unix(),
		"pwd": hashPassword,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := jwtToken.SignedString([]byte(a.Config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("cannot sign JWT: %w", err)
	}

	return signedToken, nil
}

func (a *AuthService) ValidateJWT(request *http.Request) error {

	cookie, err := request.Cookie("token")
	if err != nil {
		return errors.New("token not found")
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(a.Config.SecretKey), nil
	})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return errors.New("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return errors.New("token expired")
		}
	} else {
		return errors.New("missing exp claim")
	}

	hashFromToken, ok := claims["pwd"].(string)
	if !ok {
		return errors.New("missing password hash in token")
	}

	hash := sha256.New()
	hash.Write([]byte(a.Config.Password))
	hashPassword := hex.EncodeToString(hash.Sum(nil))

	if hashFromToken != hashPassword {
		return errors.New("invalid token: password hash mismatch")
	}

	return nil
}

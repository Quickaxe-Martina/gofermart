package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"github.com/golang-jwt/jwt/v5"
)

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID   int
	UserName string
}

// ErrNoJWTInCookie indicates that there is no JWT token in the cookie
var ErrNoJWTInCookie = errors.New("no jwt in cookie")

// ErrInvalidJWTToken indicates that the JWT token is invalid
var ErrInvalidJWTToken = errors.New("invalid jwt token")

// Имя куки, в которой хранится JWT-токен
const cookieUserJWT = "jwt_token"

// GetUser TODO
func GetUser(tokenString string, secretKey string) (storage.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return storage.User{}, err
	}

	if !token.Valid {
		return storage.User{}, ErrInvalidJWTToken
	}

	return storage.User{ID: claims.UserID, UserName: claims.UserName}, nil
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(secretKey string, tokenExp time.Duration, userID int, username string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID:   userID,
		UserName: username,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// SetTokenInCookie sets the JWT token in the HTTP cookie with the specified TTL.
func SetTokenInCookie(w http.ResponseWriter, token string, ttl time.Duration) {
	cookie := &http.Cookie{
		Name:     cookieUserJWT,
		Value:    token,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(ttl),
	}
	http.SetCookie(w, cookie)
}

// GenerateAndSetTokenInCookie TODO
func GenerateAndSetTokenInCookie(w http.ResponseWriter, secretKey string, tokenExp time.Duration, userID int, username string) error {
	token, err := BuildJWTString(secretKey, tokenExp, userID, username)
	if err != nil {
		return err
	}
	SetTokenInCookie(w, token, tokenExp)
	return nil
}

// GetUserByCookie extracts the JWT token from the request cookie,
// validates it, and returns the associated User.
func GetUserByCookie(r *http.Request, secretKey string) (storage.User, error) {
	cookie, err := r.Cookie(cookieUserJWT)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return storage.User{}, ErrNoJWTInCookie
		}
		return storage.User{}, err
	}
	user, err := GetUser(cookie.Value, secretKey)
	if err != nil {
		return storage.User{}, ErrInvalidJWTToken
	}
	return user, nil
}

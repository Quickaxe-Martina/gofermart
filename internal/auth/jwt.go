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
	UserID int
}

// ErrNoJWTInCookie indicates that there is no JWT token in the cookie
var ErrNoJWTInCookie = errors.New("no jwt in cookie")

// ErrInvalidJWTToken indicates that the JWT token is invalid
var ErrInvalidJWTToken = errors.New("invalid jwt token")

// Имя куки, в которой хранится JWT-токен
const cookieUserJWT = "jwt_token"

// GetUserID принимает JWT-токен в виде строки, парсит его и возвращает UserID из утверждений.
// Если токен недействителен или произошла ошибка при парсинге, возвращается -1.
func GetUserID(tokenString string, secretKey string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		return -1
	}

	return claims.UserID
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func BuildJWTString(secretKey string, tokenExp time.Duration, userID int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID: userID,
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
		Expires:  time.Now().Add(ttl),
	}
	http.SetCookie(w, cookie)
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
	userID := GetUserID(cookie.Value, secretKey)
	if userID == -1 {
		return storage.User{}, ErrInvalidJWTToken
	}
	return storage.User{ID: userID}, nil
}

// GetOrCreateUser retrieves the user from the request cookie or creates a new user if not found.
func GetOrCreateUser(w http.ResponseWriter, r *http.Request, store storage.Storage, secretKey string, tokenExp time.Duration) (storage.User, error) {
	user, err := GetUserByCookie(r, secretKey)
	if err != nil {
		if errors.Is(err, ErrNoJWTInCookie) || errors.Is(err, ErrInvalidJWTToken) {
			user, err = store.CreateUser(r.Context())
			if err != nil {
				return storage.User{}, err
			}
			tokenString, err := BuildJWTString(secretKey, tokenExp, user.ID)
			if err != nil {
				return storage.User{}, err
			}
			SetTokenInCookie(w, tokenString, tokenExp)
			return user, nil
		}
		return storage.User{}, err
	}
	return user, nil
}

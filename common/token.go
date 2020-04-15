package common

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// This key will be exported
var key []byte = []byte("WlAye5L1uzZq61p41A6PyhpBxsnklABk6FPAOeOXUwqWuouUEvTG8Apqkqo1uloZ")

func getClaims(username string) *jwt.StandardClaims {
	expDuration, _ := time.ParseDuration("8h")
	nbf := time.Now()
	exp := nbf.Add(expDuration)
	return &jwt.StandardClaims{
		ExpiresAt: exp.Unix(),
		NotBefore: nbf.Unix(),
		Issuer:    "MusicApp",
		Id:        username,
	}
}

// CreateToken returns a jwt signed by the key
func CreateToken(username string) (string, error) {
	claims := getClaims(username)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

// KeyToken will be used as key in the context of a request.
type KeyToken struct {
	Username, Role string
}

// GetValidateTokenMiddleWare return middleware to validate a token
func GetValidateTokenMiddleWare(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			if request.Header["Token"] == nil {
				logger.Println("[WARNING] Unauthorized request.")
				http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			token, err := jwt.Parse(request.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				logger.Printf("[WARNING] Invalid token: %v\n", err)
				http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			claims := token.Claims.(jwt.MapClaims)
			tokenContext := KeyToken{Username: claims["jti"].(string), Role: claims["iss"].(string)}
			ctx := context.WithValue(request.Context(), KeyToken{}, tokenContext)
			request = request.WithContext(ctx)
			next.ServeHTTP(response, request)
		})
	}
}

// GetIsAdminMiddleWare returns middleware that checks if a token belongs to an admin
func GetIsAdminMiddleWare(logger *log.Logger) func(http.Handler) http.Handler {
	tokenValidator := GetValidateTokenMiddleWare(logger)
	return func(next http.Handler) http.Handler {
		return tokenValidator(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			ctx := request.Context().Value(KeyToken{}).(KeyToken)
			if ctx.Role != "admin" {
				logger.Printf("[WARNING] Non admin tries to access admin content: %v\n", ctx.Username)
				http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(response, request)
		}))
	}
}

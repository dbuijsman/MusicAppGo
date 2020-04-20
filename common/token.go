package common

import (
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

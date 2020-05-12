package server

import (
	"errors"
	"fmt"
	"general/types"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// This key is currently hard-coded
var key []byte = []byte("WlAye5L1uzZq61p41A6PyhpBxsnklABk6FPAOeOXUwqWuouUEvTG8Apqkqo1uloZ")

// RoleInternal is the role for requests coming from a service
const RoleInternal string = "MusicApp"

func getClaims(id int, username, role string) *jwt.StandardClaims {
	expDuration, _ := time.ParseDuration("8h")
	nbf := time.Now()
	exp := nbf.Add(expDuration)
	return &jwt.StandardClaims{
		ExpiresAt: exp.Unix(),
		NotBefore: nbf.Unix(),
		Id:        strconv.Itoa(id),
		Issuer:    username,
		Audience:  role,
	}
}

// CreateToken returns a jwt signed by the key
func CreateToken(id int, username, role string) (string, error) {
	claims := getClaims(id, username, role)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

// CreateTokenInternalRequests creates a jwt for requests between services
func CreateTokenInternalRequests(servername string) (string, error) {
	return CreateToken(-1, servername, RoleInternal)
}

func validateToken(tokenString string) (types.Credentials, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return types.Credentials{}, errors.New("invalid signing method")
		}
		return key, nil
	})
	if err != nil || !token.Valid {
		return types.Credentials{}, fmt.Errorf("[WARNING] Invalid token: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	id, convError := strconv.Atoi(claims["jti"].(string))
	if convError != nil && claims["jti"].(string) != "" {
		return types.Credentials{}, fmt.Errorf("[WARNING] Received signed token with invalid id: %v", convError)
	}
	tokenContext := types.Credentials{ID: id, Username: claims["iss"].(string), Role: claims["aud"].(string)}
	return tokenContext, nil
}

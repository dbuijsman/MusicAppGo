package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
)

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

// OffsetMax contains the offset and the max of a request
type OffsetMax struct {
	Offset, Max int
}

// GetOffsetMaxMiddleware returns middleware for obtaining the offset and max value of the query of the request. This will also check if the values of offset and max are valid.
func GetOffsetMaxMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			queries := request.URL.Query()
			query := queries.Get("offset")
			offset, errOffset := strconv.Atoi(query)
			if query != "" && errOffset != nil {
				logger.Printf("Got request with non-numeric value for offset: %s", errOffset)
				http.Error(response, "Invalid query value.", http.StatusBadRequest)
				return
			}
			query = queries.Get("max")
			max, errMax := strconv.Atoi(query)
			if query != "" && errMax != nil {
				logger.Printf("Got request with non-numeric value for max: %s", errMax)
				http.Error(response, "Invalid query value.", http.StatusBadRequest)
				return
			}
			if query == "" || max > 20 {
				max = 20
			}
			if offset < 0 || max <= 0 {
				logger.Printf("Got request with invalid numeric values: offset=%v and max=%v\n", offset, max)
				http.Error(response, "Invalid query value.", http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(request.Context(), OffsetMax{}, OffsetMax{Offset: offset, Max: max})
			request = request.WithContext(ctx)
			next.ServeHTTP(response, request)
		})
	}
}

// GetOffsetMaxFromRequest extracts the offset and max from the request.
func GetOffsetMaxFromRequest(request *http.Request) (offset, max int, err error) {
	queries := request.URL.Query()
	query := queries.Get("offset")
	offset, err = strconv.Atoi(query)
	if query != "" && err != nil {
		errorMessage := fmt.Sprintf("Got request with invalid value for offset: %s", err)
		err = errors.New(errorMessage)
		return
	}
	query = queries.Get("max")
	max, err = strconv.Atoi(query)
	if query != "" && err != nil {
		errorMessage := fmt.Sprintf("Got request with invalid value for max: %s", err)
		err = errors.New(errorMessage)
		return
	}
	err = nil
	if query == "" {
		max = 20
	}
	return
}

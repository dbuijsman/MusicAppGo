package general

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
)

func toMiddlerWare(functionAsMiddleWare func(http.ResponseWriter, *http.Request, http.Handler)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			functionAsMiddleWare(response, request, next)
		})
	}
}

func validateToken(tokenString string) (Credentials, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return Credentials{}, errors.New("invalid signing method")
		}
		return key, nil
	})
	if err != nil || !token.Valid {
		return Credentials{}, fmt.Errorf("[WARNING] Invalid token: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	id, convError := strconv.Atoi(claims["jti"].(string))
	if convError != nil && claims["jti"].(string) != "" {
		return Credentials{}, fmt.Errorf("[WARNING] Received signed token with invalid id: %v", convError)
	}
	tokenContext := Credentials{ID: id, Username: claims["iss"].(string), Role: claims["aud"].(string)}
	return tokenContext, nil
}

// GetValidateTokenMiddleWare return middleware to validate a token
func GetValidateTokenMiddleWare(logger *log.Logger) func(http.Handler) http.Handler {
	return toMiddlerWare(func(response http.ResponseWriter, request *http.Request, next http.Handler) {
		if request.Header["Token"] == nil {
			logger.Println("[WARNING] Unauthorized request\n")
			http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		token, err := validateToken(request.Header["Token"][0])
		if err != nil {
			logger.Printf("[WARNING] Request with invalid token: %s\n", err)
			http.Error(response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(request.Context(), Credentials{}, token)
		request = request.WithContext(ctx)
		next.ServeHTTP(response, request)
	})
}

// GetAddTokenToContextMiddleware returns middleware to validate a token but it won't deny a request
func GetAddTokenToContextMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return toMiddlerWare(func(response http.ResponseWriter, request *http.Request, next http.Handler) {
		if request.Header["Token"] == nil {
			next.ServeHTTP(response, request)
			return
		}
		token, err := validateToken(request.Header["Token"][0])
		if err != nil {
			logger.Printf("[WARNING] request with invalid token: %s\n", err)
			next.ServeHTTP(response, request)
			return
		}
		ctx := context.WithValue(request.Context(), Credentials{}, token)
		request = request.WithContext(ctx)
		next.ServeHTTP(response, request)
	})
}

// GetIsAdminMiddleware returns middleware that checks if a token belongs to an admin
func GetIsAdminMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return GetTokenMiddleWareForSpecificRole(logger, "admin")
}

// GetInternalRequestMiddleware returns middleware that checks if a token belongs to a service within this application
func GetInternalRequestMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return GetTokenMiddleWareForSpecificRole(logger, RoleInternal)
}

// GetTokenMiddleWareForSpecificRole returns middleware that checks if a token belongs to the given role
func GetTokenMiddleWareForSpecificRole(logger *log.Logger, role string) func(http.Handler) http.Handler {
	tokenValidator := GetValidateTokenMiddleWare(logger)
	return func(next http.Handler) http.Handler {
		return tokenValidator(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			ctx := request.Context().Value(Credentials{}).(Credentials)
			if ctx.Role != role {
				logger.Printf("[WARNING] Non-%v tries to access %v content: %v\n", role, role, ctx.Username)
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
	return toMiddlerWare(func(response http.ResponseWriter, request *http.Request, next http.Handler) {
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

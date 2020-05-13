package server

import (
	"context"
	"general/types"
	"log"
	"net/http"
	"strconv"
)

func toMiddlerWare(functionAsMiddleWare func(http.ResponseWriter, *http.Request, http.Handler)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			functionAsMiddleWare(response, request, next)
		})
	}
}

// GetValidateTokenMiddleWare return middleware to validate a token
func GetValidateTokenMiddleWare(logger *log.Logger) func(http.Handler) http.Handler {
	return toMiddlerWare(func(response http.ResponseWriter, request *http.Request, next http.Handler) {
		if request.Header["Token"] == nil {
			logger.Printf("[WARNING] Unauthorized request at %v\n", request.URL.Path)
			SendError(response, http.StatusUnauthorized)
			return
		}
		token, err := validateToken(request.Header["Token"][0])
		if err != nil {
			logger.Printf("[WARNING] Unauthorized request at %v with invalid token: %s\n", request.URL.Path, err)
			SendError(response, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(request.Context(), types.Credentials{}, token)
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
			logger.Printf("[WARNING] Request at %v with invalid token: %s\n", request.URL.Path, err)
			next.ServeHTTP(response, request)
			return
		}
		ctx := context.WithValue(request.Context(), types.Credentials{}, token)
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
			ctx := request.Context().Value(types.Credentials{}).(types.Credentials)
			if ctx.Role != role {
				logger.Printf("[WARNING] Non-%v tries to access %v content at %v: %v\n", role, role, request.URL.Path, ctx.Username)
				SendError(response, http.StatusUnauthorized)
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

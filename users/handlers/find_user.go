package handlers

import (
	"common"
	"net/http"
)

// GetRole returns the username and role that belongs to the token (This was a test version and will be changed)
func (handler *UserHandler) GetRole(response http.ResponseWriter, request *http.Request) {
	username := request.Context().Value(common.KeyToken{}).(common.KeyToken).Username
	if username == "" {
		badRequests.Inc()
		handler.Logger.Printf("Got request with invalid token: %v\n", request.Header["Token"][0])
		http.Error(response, "Bad request", http.StatusBadRequest)
		return
	}
	user, err := handler.db.FindUser(username)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to find user in database: %v\n", err.Error())
		http.Error(response, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user.Username == "" {
		handler.Logger.Printf("[ERROR] Can't find user %v from token in database users\n", username)
		http.Error(response, "User not found.", http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	common.ToJSON(user, response)
}

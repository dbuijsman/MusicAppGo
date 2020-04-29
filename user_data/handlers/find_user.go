package handlers

import (
	"general"
	"net/http"
)

// GetRole returns the username and role that belongs to the token (This was a test version and will be changed)
func (handler *UserHandler) GetRole(response http.ResponseWriter, request *http.Request) {
	username := request.Context().Value(general.Credentials{}).(general.Credentials).Username
	user, err := handler.db.FindUser(username)
	if err != nil {
		if err.(general.DBError).ErrorCode != general.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to find user in database: %v\n", err.Error())
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("[ERROR] Can't find user %v from token in database users\n", username)
		general.SendError(response, http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	general.WriteToJSON(user, response)
}

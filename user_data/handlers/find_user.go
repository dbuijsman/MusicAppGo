package handlers

import (
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"
	"net/http"
)

// GetRole returns the username and role that belongs to the token (This was a test version and will be changed)
func (handler *UserHandler) GetRole(response http.ResponseWriter, request *http.Request) {
	username := request.Context().Value(types.Credentials{}).(types.Credentials).Username
	user, err := handler.db.FindUser(username)
	if err != nil {
		if err.(dberror.DBError).ErrorCode != dberror.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to find user in database: %v\n", err.Error())
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("[ERROR] Can't find user %v from token in database users\n", username)
		server.SendError(response, http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	convert.WriteToJSON(user, response)
}

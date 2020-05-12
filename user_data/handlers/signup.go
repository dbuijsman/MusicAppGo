package handlers

import (
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"
	"net/http"
)

// SignUp handles the request to add a new user to the database.
func (handler *UserHandler) SignUp(response http.ResponseWriter, request *http.Request) {
	var creds ClientCredentials
	if err := convert.ReadFromJSON(&creds, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid signup request: %v\n", err)
		server.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for new user: %v\n", creds.Username)
	userID, err := handler.db.SignUp(creds.Username, creds.Password)
	if err != nil {
		if err.(dberror.DBError).ErrorCode == dberror.DuplicateEntry {
			handler.Logger.Printf("Duplicate username: %v\n", creds.Username)
			http.Error(response, "This username already exists", http.StatusUnprocessableEntity)
			return
		}
		failedSignUps.Inc()
		handler.Logger.Printf("[ERROR] Failed to save credentials for new user %v in database: %s\n", creds.Username, err)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Succesfully added new user: %v\n", creds.Username)
	newUser := types.Credentials{ID: userID, Username: creds.Username}
	go func(credentials types.Credentials) {
		msg, _ := convert.ToJSONBytes(&newUser)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to convert user with ID:%v, username: %v to bytes: %v\n", credentials.ID, credentials.Username, err)
			return
		}
		handler.SendMessage("newUser", msg)
	}(newUser)
	succesSignUps.Inc()
	handler.sendToken(newUser, response)
}

package handlers

import (
	"general"
	"net/http"
)

// Login checks if the given credentials conincide with credentials in the database
func (handler *UserHandler) Login(response http.ResponseWriter, request *http.Request) {
	var creds ClientCredentials
	if err := general.ReadFromJSON(&creds, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("[ERROR] Invalid login request: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received login call for user: %v\n", creds.Username)
	result, err := handler.db.Login(creds.Username, creds.Password)
	if err != nil {
		if err.(general.DBError).ErrorCode != general.InvalidInput {
			failServerLogin.Inc()
			handler.Logger.Printf("[ERROR] Failed to retrieve credentials from database: %v\n", err.Error())
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("User %v sends incorrect credentials\n", creds.Username)
		http.Error(response, "Username and password do not match.", http.StatusUnauthorized)
		return
	}
	handler.Logger.Printf("User %v succesfully logged in\n", creds.Username)
	go func(user general.Credentials) {
		msg, _ := general.ToJSONBytes(user)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to convert user with ID:%v, username: %v to bytes: %v\n", user.ID, user.Username, err)
			return
		}
		handler.SendMessage("login", msg)
	}(result)
	succesLogin.Inc()
	handler.sendToken(result, response)
}

package handlers

import (
	"MusicAppGo/common"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	succesSignUps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_signup_total",
		Help: "The total number of new users",
	})
)

var (
	failedSignUps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_signup_denied_total",
		Help: "The total number of failed requests to add a new user",
	})
)

// SignUp handles the request to add a new user to the database.
func (handler *UserHandler) SignUp(response http.ResponseWriter, request *http.Request) {
	var creds Credentials
	if err := common.FromJSON(&creds, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid signup request: %v\n", err)
		http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err := handler.db.InsertUser(creds.Username, creds.Password); err != nil {
		if err.(common.DBError).ErrorCode == common.DuplicateEntry {
			handler.Logger.Printf("Duplicate username: %v\n", creds.Username)
			http.Error(response, "This username already exists", http.StatusUnprocessableEntity)
			return
		}
		failedSignUps.Inc()
		handler.Logger.Printf("[ERROR] Failed to save credentials in database: %v\n", err.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Succesfully added new user %v\n", creds.Username)
	succesSignUps.Inc()
	token, err := common.CreateToken(creds.Username)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to create valid jwt: %v\n", err.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "text/plain;charset=utf-8")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(token))
}

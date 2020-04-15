package handlers

import (
	"common"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	succesLoggin = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_login_total",
		Help: "The total number of succes logins",
	})
)

var (
	failServerLogin = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_login_server_error_total",
		Help: "The total number of failed requests to confirm a login attempt",
	})
)

// Login checks if the given credentials conincide with credentials in the database
func (handler *UserHandler) Login(response http.ResponseWriter, request *http.Request) {
	var creds Credentials
	if err := common.FromJSON(&creds, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("[ERROR] Invalid login request: %v\n", err)
		http.Error(response, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	result, err := handler.db.Login(creds.Username, creds.Password)
	if err != nil {
		failServerLogin.Inc()
		handler.Logger.Printf("[ERROR] Failed to retrieve credentials from database: %v\n", err.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if result == false {
		http.Error(response, "Username and password do not match.", http.StatusUnauthorized)
		return
	}
	handler.Logger.Printf("User %v succesfully logged in\n", creds.Username)
	succesLoggin.Inc()
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

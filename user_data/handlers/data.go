package handlers

import (
	"general"
	"log"
	"net/http"
	"user_data/database"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// UserHandler consists of a logger and a database
type UserHandler struct {
	Logger      *log.Logger
	db          database.Database
	SendMessage func(string, []byte)
}

//NewUserHandler returns a UserHandler
func NewUserHandler(l *log.Logger, db database.Database, sendMessage func(string, []byte) error) *UserHandler {
	return &UserHandler{Logger: l, db: db, SendMessage: func(topic string, message []byte) {
		if err := sendMessage(topic, message); err != nil {
			l.Printf("Topic %v: Can't send message %s: %v\n", topic, message, err)
			return
		}
		l.Printf("Topic %v: Send message: %s\n", topic, message)
	}}
}

// ClientCredentials contains the credentials that were send from the client
type ClientCredentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// NewClientCredentials returns ClientCredentials with the given data
func NewClientCredentials(username, password string) ClientCredentials {
	return ClientCredentials{Username: username, Password: password}
}

func (handler *UserHandler) sendToken(creds general.Credentials, response http.ResponseWriter) {
	if creds.Role == "" {
		creds.Role = "user"
	}
	token, err := general.CreateToken(creds.ID, creds.Username, creds.Role)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to create valid jwt: %v\n", err.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Succesfully created token for user: %v\n", creds.Username)
	response.Header().Set("Content-Type", "text/plain;charset=utf-8")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(token))
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
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

var (
	succesLogin = promauto.NewCounter(prometheus.CounterOpts{
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

package handlers

import (
	"errors"
	"general"
	"log"
	"net/http"
	"user_data/database"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const servername string = "users"

// NewUserServer returns a new server for userdata and a function that starts up the server
func NewUserServer(handler *UserHandler, broker *kafka.Broker, servername, port string) (server *http.Server, start func()) {
	server, _, start = general.NewServer(servername, port, initRoutes(handler), broker, nil, handler.Logger)
	return
}

// initRoutes will returns a router with the necessary routes registered to it.
func initRoutes(users *UserHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/signup", users.SignUp)
	postRouter.HandleFunc("/login", users.Login)

	validateRouter := router.PathPrefix("/validate").Subrouter()
	validateRouter.HandleFunc("/", users.GetRole)
	validateRouter.Use(general.GetValidateTokenMiddleWare(users.Logger))
	return router
}

// UserHandler consists of a logger and a database
type UserHandler struct {
	Logger      *log.Logger
	db          database.Database
	SendMessage func(string, []byte)
}

//NewUserHandler returns a UserHandler. It returns an error if sendMessage is nil.
func NewUserHandler(logger *log.Logger, db database.Database, sendMessage func(string, []byte) error) (*UserHandler, error) {
	if sendMessage == nil {
		return nil, errors.New("sendMessage can't be nil")
	}
	return &UserHandler{Logger: logger, db: db, SendMessage: func(topic string, message []byte) {
		if err := sendMessage(topic, message); err != nil {
			logger.Printf("Topic %v: Can't send message %s: %v\n", topic, message, err)
			return
		}
		logger.Printf("Topic %v: Send message: %s\n", topic, message)
	}}, nil
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

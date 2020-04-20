package handlers

import (
	"log"
	"users/database"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Credentials consists of user credentials
type Credentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

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
			l.Printf("Topic %v: Can't send message %v: %v\n", topic, message, err)
		}
	}}
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

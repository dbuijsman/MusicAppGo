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
	Logger *log.Logger
	db     database.Database
}

//NewUserHandler returns a UserHandler
func NewUserHandler(l *log.Logger, db database.Database) *UserHandler {
	return &UserHandler{Logger: l, db: db}
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

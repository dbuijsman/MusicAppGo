package handlers

import (
	"likes/database"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const portDiscography string = ":9002"

// LikesHandler consists of a logger and a database
type LikesHandler struct {
	Logger     *log.Logger
	db         database.Database
	GETRequest func(string) (*http.Response, error)
}

//NewLikesHandler returns a MusicHandler
func NewLikesHandler(l *log.Logger, db database.Database, get func(string) (*http.Response, error)) *LikesHandler {
	return &LikesHandler{Logger: l, db: db, GETRequest: get}
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

var (
	failureGetRequest = promauto.NewCounter(prometheus.CounterOpts{
		Name: "likes_failed_get_request_total",
		Help: "The total number of failed requests to find preferences of an user that satisfies the requirements",
	})
)

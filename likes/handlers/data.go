package handlers

import (
	"general"
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

// DefealtGETRequest can be used as value for LikesHandler.GETRequest
func DefealtGETRequest(servername string) (func(string) (*http.Response, error), error) {
	tokenString, errToken := general.CreateTokenInternalRequests(servername)
	token := &tokenString
	client := http.Client{}
	if errToken != nil {
		return nil, errToken
	}
	return func(url string) (*http.Response, error) {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Token", *token)
		resp, respError := client.Do(request)
		if resp.StatusCode == http.StatusUnauthorized {
			*token, errToken = general.CreateTokenInternalRequests(servername)
			request, err = http.NewRequest("GET", url, nil)
			request.Header.Add("Token", *token)
			return client.Do(request)
		}
		return resp, respError
	}, nil
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "users_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

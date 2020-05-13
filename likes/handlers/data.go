package handlers

import (
	"errors"
	"general/env"
	"general/server"
	"likes/database"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var nameDiscography = env.SetString("DEP_DISCOGRAPHY_NAME", false, "likes", "Name of the discography service. Needed for adding missing data")
var addressDiscography string

// NewLikesServer returns a new server for likes and dislikes and a function that starts up the server.
// If broker is nil, then there will be no messages consumed.
func NewLikesServer(handler *LikesHandler, broker *kafka.Broker, servername, host string, port int) (newServer *http.Server, start func()) {
	var startConsumer func()
	if broker != nil {
		startConsumer = func() {
			handler.StartConsuming(broker)
		}
	}
	s, channel, startServer := server.NewServer(servername, host, port, initRoutes(handler), broker, startConsumer, handler.Logger)
	newServer = s
	start = func() {
		go func() {
			for service := range channel {
				if service.Name == *nameDiscography {
					addressDiscography = service.Address
				}
			}
		}()
		startServer()
	}
	return
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *LikesHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	clientR := router.PathPrefix("/api").Subrouter()
	clientR.Use(server.GetValidateTokenMiddleWare(handler.Logger))

	getR := clientR.Methods(http.MethodGet).Subrouter()
	getR.Use(server.GetOffsetMaxMiddleware(handler.Logger))
	getR.PathPrefix("/like").HandlerFunc(handler.GetLikes)
	getR.PathPrefix("/dislike").HandlerFunc(handler.GetDislikes)

	likesR := clientR.PathPrefix("/like").Subrouter()
	likesR.Methods(http.MethodPost).HandlerFunc(handler.AddLike)
	likesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveLike)

	dislikesR := clientR.PathPrefix("/dislike").Subrouter()
	dislikesR.Methods(http.MethodPost).HandlerFunc(handler.AddDislike)
	dislikesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveDislike)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(server.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/preference/{user}/{artist}").HandlerFunc(handler.GetPreferencesOfArtist)
	return router
}

// LikesHandler consists of a logger and a database
type LikesHandler struct {
	Logger     *log.Logger
	db         database.Database
	GETRequest func(string) (*http.Response, error)
}

//NewLikesHandler returns a MusicHandler. Get can't be nil
func NewLikesHandler(logger *log.Logger, db database.Database, get func(string) (*http.Response, error)) (*LikesHandler, error) {
	if get == nil {
		return nil, errors.New("get can't be nil")
	}
	return &LikesHandler{Logger: logger, db: db, GETRequest: get}, nil
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

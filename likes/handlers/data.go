package handlers

import (
	"general"
	"likes/database"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const servername string = "likes"

var portDiscography string

// NewLikesServer returns a new server for likes and dislikes and a function that starts up the server.
// If broker is nil, then there will be no messages consumed.
func NewLikesServer(handler *LikesHandler, broker *kafka.Broker, servername, port string) (server *http.Server, start func()) {
	var startConsumer func()
	if broker != nil {
		startConsumer = func() {
			handler.StartConsuming(broker)
		}
	}
	s, channel, startServer := general.NewServer(servername, port, initRoutes(handler), broker, startConsumer, handler.Logger)
	server = s
	start = func() {
		go func() {
			for service := range channel {
				if service.Name == "discography" {
					portDiscography = service.Address
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
	clientR.Use(general.GetValidateTokenMiddleWare(handler.Logger))

	getR := clientR.Methods(http.MethodGet).Subrouter()
	getR.Use(general.GetOffsetMaxMiddleware(handler.Logger))
	getR.PathPrefix("/like").HandlerFunc(handler.GetLikes)
	getR.PathPrefix("/dislike").HandlerFunc(handler.GetDislikes)

	likesR := clientR.PathPrefix("/like").Subrouter()
	likesR.Methods(http.MethodPost).HandlerFunc(handler.AddLike)
	likesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveLike)

	dislikesR := clientR.PathPrefix("/dislike").Subrouter()
	dislikesR.Methods(http.MethodPost).HandlerFunc(handler.AddDislike)
	dislikesR.Methods(http.MethodDelete).HandlerFunc(handler.RemoveDislike)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(general.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/preference/{user}/{artist}").HandlerFunc(handler.GetPreferencesOfArtist)
	return router
}

// LikesHandler consists of a logger and a database
type LikesHandler struct {
	Logger     *log.Logger
	db         database.Database
	GETRequest func(string) (*http.Response, error)
}

//NewLikesHandler returns a MusicHandler.
// If get is nil, then DefaultGETRequest will be used with the default servername
func NewLikesHandler(logger *log.Logger, db database.Database, get func(string) (*http.Response, error)) *LikesHandler {
	if get == nil {
		var err error
		get, err = general.GetInternalGETRequest(servername)
		if err != nil {
			logger.Fatalf("Can't create a client for sending get requests: %s\n", err)
		}
	}
	return &LikesHandler{Logger: logger, db: db, GETRequest: get}
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

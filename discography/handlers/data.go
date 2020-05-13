package handlers

import (
	"MusicAppGo/general/env"
	"discography/database"
	"errors"
	"general/server"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/optiopay/kafka/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var nameLikes = env.SetString("DEP_LIKES_NAME", false, "likes", "Name of the likes service. Needed for adding preferences to songs")
var addressLikes string

// NewMusicServer returns a new server for music and a function that starts up the server
func NewMusicServer(handler *MusicHandler, broker *kafka.Broker, servername, host string, port int) (newServer *http.Server, start func()) {
	s, channel, startServer := server.NewServer(servername, host, port, initRoutes(handler), broker, nil, handler.Logger)
	newServer = s
	start = func() {
		go func() {
			for service := range channel {
				if service.Name == *nameLikes {
					addressLikes = service.Address
				}
			}
		}()
		startServer()
	}
	return
}

// initRoutes returns a router which can handle all the requests for this microservice
func initRoutes(handler *MusicHandler) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	getR := router.PathPrefix("/api").Methods(http.MethodGet).Subrouter()
	getR.Use(server.GetAddTokenToContextMiddleware(handler.Logger))
	getR.Use(server.GetOffsetMaxMiddleware(handler.Logger))
	getR.Path("/artists/{firstLetter}").HandlerFunc(handler.ArtistStartingWith)
	getR.Path("/artist/{artist}").HandlerFunc(handler.SongsFromArtist)

	adminR := router.PathPrefix("/admin").Methods(http.MethodPost).Subrouter()
	adminR.Use(server.GetIsAdminMiddleware(handler.Logger))
	adminR.Path("/artist").HandlerFunc(handler.AddArtistHandler)
	adminR.Path("/song").HandlerFunc(handler.AddSongHandler)

	internalR := router.PathPrefix("/intern").Methods(http.MethodGet).Subrouter()
	internalR.Use(server.GetInternalRequestMiddleware(handler.Logger))
	internalR.Path("/artist/{id}").HandlerFunc(handler.FindArtistByID)
	internalR.Path("/song/{id}").HandlerFunc(handler.FindSongByID)
	return router
}

// MusicHandler consists of a logger and a database
type MusicHandler struct {
	Logger      *log.Logger
	db          database.Database
	SendMessage func(string, []byte)
	GETRequest  func(string) (*http.Response, error)
}

//NewMusicHandler returns a MusicHandler.
// If get is nil, then DefaultGETRequest will be used with the default servername
// Returns error if error is nil or if DefaultGetRequest returns an error
func NewMusicHandler(logger *log.Logger, db database.Database, sendMessage func(string, []byte) error, get func(string) (*http.Response, error)) (*MusicHandler, error) {
	if sendMessage == nil {
		return nil, errors.New("sendMessage can't be nil")
	}
	if get == nil {
		return nil, errors.New("get can't be nil")
	}
	return &MusicHandler{Logger: logger, db: db, GETRequest: get,
		SendMessage: func(topic string, message []byte) {
			if err := sendMessage(topic, message); err != nil {
				logger.Printf("Topic %v: Can't send message %s: %v\n", topic, message, err)
				return
			}
			logger.Printf("Topic %v: Send message: %s\n", topic, message)
		},
	}, nil
}

// ClientArtist is the form that is used in posting a new artist from the client side
type ClientArtist struct {
	Artist      string `json:"artist" validate:"required"`
	LinkSpotify string `json:"linkSpotify" validate:"required"`
}

// NewClientArtist returns a ClientArtist containing the given data
func NewClientArtist(artist, linkSpotify string) ClientArtist {
	return ClientArtist{Artist: artist, LinkSpotify: linkSpotify}
}

// ClientSong is the form that is used in posting a new song from the client side
type ClientSong struct {
	Artists []string `json:"artists" validate:"required"`
	Name    string   `json:"song" validate:"required"`
}

// NewClientSong returns a ClientArtist containing the given data
func NewClientSong(song string, artists ...string) ClientSong {
	return ClientSong{Artists: artists, Name: song}
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "music_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

var (
	failureSearchRequest = promauto.NewCounter(prometheus.CounterOpts{
		Name: "discography_failed_search_request_total",
		Help: "The total number of failed requests to find artists or songs that satisfies the requirements",
	})
)

var (
	succesNewArtist = promauto.NewCounter(prometheus.CounterOpts{
		Name: "admin_new_artist_total",
		Help: "The total number of succesfull requests to add a new artist to the database",
	})
)

var (
	failedNewArtist = promauto.NewCounter(prometheus.CounterOpts{
		Name: "admin_new_artist_denied_total",
		Help: "The total number of failed requests to add a new artist to the database",
	})
)
var (
	succesNewSong = promauto.NewCounter(prometheus.CounterOpts{
		Name: "admin_new_song_total",
		Help: "The total number of succesfull requests to add a new song to the database",
	})
)

var (
	failedNewSong = promauto.NewCounter(prometheus.CounterOpts{
		Name: "admin_new_song_denied_total",
		Help: "The total number of failed requests to add a new song to the database",
	})
)

package handlers

import (
	"discography/database"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MusicHandler consists of a logger and a database
type MusicHandler struct {
	Logger      *log.Logger
	db          database.Database
	SendMessage func(string, []byte)
}

//NewMusicHandler returns a MusicHandler
func NewMusicHandler(l *log.Logger, db database.Database, sendMessage func(string, []byte) error) *MusicHandler {
	return &MusicHandler{Logger: l, db: db, SendMessage: func(topic string, message []byte) {
		if err := sendMessage(topic, message); err != nil {
			l.Printf("Topic %v: Can't send message %s: %v\n", topic, message, err)
			return
		}
		l.Printf("Topic %v: Send message: %s\n", topic, message)
	}}
}

// ClientArtist is the form that is used in posting a new artist from the client side
type ClientArtist struct {
	Artist      string `json:"artist" validate:"required"`
	LinkSpotify string `json:"linkSpotify"`
}

// NewClientArtist returns a ClientArtist containing the given data
func NewClientArtist(artist, linkSpotify string) ClientArtist {
	return ClientArtist{Artist: artist, LinkSpotify: linkSpotify}
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

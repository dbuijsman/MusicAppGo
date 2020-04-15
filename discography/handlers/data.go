package handlers

import (
	"discography/database"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MusicHandler consists of a logger and a database
type MusicHandler struct {
	Logger *log.Logger
	db     database.Database
}

//NewMusicHandler returns a MusicHandler
func NewMusicHandler(l *log.Logger, db database.Database) *MusicHandler {
	return &MusicHandler{Logger: l, db: db}
}

var (
	badRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "music_badRequests_total",
		Help: "The total number of bad requests send to the users server",
	})
)

// NewArtist will be the form of a new artist that will be added to the database
type NewArtist struct {
	Name        string `json:"name" validate:"required"`
	LinkSpotify string `json:"-"`
}

// MultipleArtists contains data of artists and a boolean to indicate if there are more results
type MultipleArtists struct {
	Music   []database.RowArtistDB `json: "music"`
	HasNext bool                   `json: "hasNext"`
}

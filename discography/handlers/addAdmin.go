package handlers

import (
	"MusicAppGo/common"
	"discography/database"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

// AddArtist will add a new artist to the database
func (handler *MusicHandler) AddArtist(response http.ResponseWriter, request *http.Request) {
	var newArtist NewArtist
	if err := common.FromJSON(&newArtist, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request to add a new artist: %v\n", err)
		http.Error(response, "Data received in incorrect format.", http.StatusBadRequest)
		return
	}
	artist, prefix := seperatePrefix(newArtist.Name)
	if _, err := handler.db.AddArtist(artist, prefix, newArtist.LinkSpotify); err != nil {
		if err.(common.DBError).ErrorCode == common.DuplicateEntry {
			handler.Logger.Printf("Duplicate artist: %v\n", artist)
			http.Error(response, "This artist is already in the database", http.StatusUnprocessableEntity)
			return
		}
		failedNewArtist.Inc()
		handler.Logger.Printf("[ERROR] Failed to save artist in database: %v\n", err.Error())
		http.Error(response, "Internal server error", http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Succesfully added new artist %v\n", artist)
	succesNewArtist.Inc()
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

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

// AddSong adds a song to the database. It will returns the new song in a SongDB struct
func (handler *MusicHandler) AddSong(song string, artists ...string) (database.SongDB, error) {
	if song == "" {
		handler.Logger.Printf("Received call to add a song without a song")
		return database.SongDB{}, common.GetDBError("Missing song", common.IncompleteInput)
	}
	if len(artists) == 0 {
		handler.Logger.Printf("Received call to add a song without an artist")
		return database.SongDB{}, common.GetDBError("Missing artists", common.IncompleteInput)
	}
	contributingArtists := make([]database.RowArtistDB, 0, len(artists))
	for _, artistName := range artists { // Adding all artists to a []RowArtistDB
		artist, err := handler.db.FindArtist(artistName)
		if err != nil {
			handler.Logger.Printf("Can't find artist %v while adding a new song\n", artist)
			newArtist, prefix := seperatePrefix(artistName)
			artist, err = handler.db.AddArtist(newArtist, prefix, "")
			if err != nil {
				failedNewSong.Inc()
				handler.Logger.Printf("[ERROR] Can't find or add artist %v: %v\n", artist, err)
				return database.SongDB{}, err
			}
			handler.Logger.Printf("Succesfully added new artist %v\n", artist)
		}
		contributingArtists = append(contributingArtists, artist)
	}
	// Check if the song already exists
	if existingSong, err := handler.db.FindSong(contributingArtists[0].Artist, song); err == nil || err.(common.DBError).ErrorCode != common.NotFoundError {
		if err == nil { // The song already exists
			handler.Logger.Printf("Trying to add %v - %v but this song already exists (ID %v)\n", contributingArtists[0].Artist, song, existingSong.ID)
			return database.SongDB{}, common.GetDBError("Duplicate entry", common.DuplicateEntry)
		}
		failedNewSong.Inc()
		handler.Logger.Printf("[ERROR] Failed to add new song %v - %v to database: %v\n", contributingArtists[0].Artist, song, err)
		return database.SongDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	newSong, err := handler.db.AddSong(song, contributingArtists)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to add new song %v - %v to database: %v\n", contributingArtists[0].Artist, song, err)
		failedNewSong.Inc()
	} else {
		handler.Logger.Printf("Succsefully added new song %v - %v\n", contributingArtists[0].Artist, song)
		succesNewSong.Inc()
	}
	return newSong, err
}

func seperatePrefix(name string) (artist, prefix string) {
	if len(name) < 4 {
		artist = name
		return
	}
	arrayPrefixes := []string{"A ", "An ", "The "}
	for _, entry := range arrayPrefixes {
		if name[0:len(entry)] == entry {
			prefix = strings.Trim(entry, " ")
			artist = name[len(entry):]
			return
		}
	}
	artist = name
	return
}

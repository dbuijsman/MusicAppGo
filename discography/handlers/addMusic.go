package handlers

import (
	"general"
	"net/http"
	"strings"
)

// AddArtist will add a new artist to the database
func (handler *MusicHandler) AddArtist(response http.ResponseWriter, request *http.Request) {
	var newArtist ClientArtist
	if err := general.ReadFromJSON(&newArtist, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request to add a new artist: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	artist, prefix := seperatePrefix(newArtist.Artist)
	handler.Logger.Printf("Received call for adding new artist %v, %v with link %v\n", artist, prefix, newArtist.LinkSpotify)
	if _, err := handler.AddNewArtist(artist, prefix, newArtist.LinkSpotify); err != nil {
		if err.(general.DBError).ErrorCode != general.DuplicateEntry {
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		http.Error(response, "This artist already exists", http.StatusUnprocessableEntity)
		return
	}
	succesNewArtist.Inc()
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

// AddNewArtist adds a new artist to the database. It returns an error when
func (handler *MusicHandler) AddNewArtist(artist, prefix, linkSpotify string) (general.Artist, error) {
	handler.Logger.Printf("Trying to add %v, %v to DB\n", artist, prefix)
	newArtist, err := handler.db.AddArtist(artist, prefix, linkSpotify)
	if err != nil {
		if err.(general.DBError).ErrorCode != general.DuplicateEntry {
			failedNewArtist.Inc()
			handler.Logger.Printf("[ERROR] Failed to add %v, %v due to: %s\n", artist, prefix, err)
			return general.Artist{}, general.ErrorToUnknownDBError(err)
		}
		handler.Logger.Printf("Artist %v, %v already exists\n", artist, prefix)
		return general.Artist{}, general.GetDBError("This artist is already in the database", general.DuplicateEntry)
	}
	go func(handler *MusicHandler, newArtist general.Artist) {
		msg, err := general.ToJSONBytes(newArtist)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to convert %v, %v to bytes: %v\n", newArtist.Name, newArtist.Prefix, err)
			return
		}
		handler.SendMessage("newArtist", msg)
	}(handler, newArtist)
	handler.Logger.Printf("Succesfully added new artist %v\n", artist)
	return newArtist, nil
}

// AddSongHandler is the handler used for adding a new song to the database
func (handler *MusicHandler) AddSongHandler(response http.ResponseWriter, request *http.Request) {
	var newArtist ClientSong
	if err := general.ReadFromJSON(&newArtist, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request to add a new artist: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for adding new song %v - %v with link %v\n", newArtist.Artists, newArtist.Name)
	if _, err := handler.AddSong(newArtist.Name, newArtist.Artists...); err != nil {
		if err.(general.DBError).ErrorCode != general.DuplicateEntry {
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		http.Error(response, "This artist already exists", http.StatusUnprocessableEntity)
		return
	}
	succesNewArtist.Inc()
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))

}

// AddSong adds a song to the database. It will returns the new song in a SongDB struct
func (handler *MusicHandler) AddSong(song string, artists ...string) (general.Song, error) {
	if song == "" {
		handler.Logger.Printf("Can't add song without a name: %v - %v\n", artists, song)
		return general.Song{}, general.GetDBError("Missing song", general.InvalidInput)
	}
	if len(artists) == 0 {
		handler.Logger.Printf("Can't add song without artists: %v - %v\n", artists, song)
		return general.Song{}, general.GetDBError("Missing artists", general.InvalidInput)
	}
	handler.Logger.Printf("Trying to add %v - %v\n", artists, song)
	// For every given artist we are going to find the same artist from the DB or add a new artist to the DB
	contributingArtists := make([]general.Artist, 0, len(artists))
	for _, artistName := range artists {
		name, prefix := seperatePrefix(artistName)
		artist, err := handler.db.FindArtistByName(name)
		if err != nil {
			if err.(general.DBError).ErrorCode != general.NotFoundError {
				handler.Logger.Printf("[ERROR] Failed to search for artist %v due to: %s/n", name, err)
				return general.Song{}, err
			}
			handler.Logger.Printf("Can't find artist %v while adding a new song\n", name)
			artist, err = handler.AddNewArtist(name, prefix, "")
			if err != nil {
				failedNewSong.Inc()
				handler.Logger.Printf("[ERROR] Can't find or add artist %v: %v\n", artist, err)
				return general.Song{}, general.ErrorToUnknownDBError(err)
			}
		}
		contributingArtists = append(contributingArtists, artist)
	}
	handler.Logger.Printf("Found all artists belonging to %v - %v\n", artists, song)
	newSong, err := handler.db.AddSong(song, contributingArtists)
	if err != nil {
		if err.(general.DBError).ErrorCode != general.DuplicateEntry {
			handler.Logger.Printf("[ERROR] Failed to add new song %v - %v to database: %v\n", artists, song, err)
			failedNewSong.Inc()
			return general.Song{}, general.ErrorToUnknownDBError(err)
		}
		handler.Logger.Printf("Trying to add %v - %v but this song already exists\n", artists, song)
		return general.Song{}, general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	// Sending message to topic newSong
	go func(handler *MusicHandler, newSong general.Song) {
		msg, err := general.ToJSONBytes(newSong)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to convert %v - %v to bytes: %v\n", newSong.Artists, newSong.Name, err)
			return
		}
		handler.SendMessage("newSong", msg)
	}(handler, newSong)

	handler.Logger.Printf("Succesfully added new song %v - %v\n", contributingArtists, song)
	succesNewSong.Inc()
	return newSong, nil
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

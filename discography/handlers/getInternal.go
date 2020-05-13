package handlers

import (
	"general/convert"
	"general/dberror"
	"general/server"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// FindSongByID returns the song that belongs to the given ID.
func (handler *MusicHandler) FindSongByID(response http.ResponseWriter, request *http.Request) {
	songIDstring := mux.Vars(request)["id"]
	songID, err := strconv.Atoi(songIDstring)
	if err != nil {
		handler.Logger.Printf("FindSongByID: Received request with invalid id %v results in: %s\n", songIDstring, err)
		server.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received request for finding song #%v\n", songID)
	song, searchErr := handler.db.FindSongByID(songID)
	if searchErr != nil {
		if searchErr.(dberror.DBError).ErrorCode != dberror.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to search DB for song #%v due to: %s\n", songID, searchErr)
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("Song #%v doesn't exists\n", songID)
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found song #%v: %v - %v\n", songID, song.Artists, song.Name)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = convert.WriteToJSON(&song, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

// FindArtistByID returns the artist that belongs to the given ID.
func (handler *MusicHandler) FindArtistByID(response http.ResponseWriter, request *http.Request) {
	artistIDstring := mux.Vars(request)["id"]
	artistID, err := strconv.Atoi(artistIDstring)
	if err != nil {
		handler.Logger.Printf("Received request with invalid id %v results in: %s\n", artistIDstring, err)
		server.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received request for finding artist #%v\n", artistID)
	artist, searchErr := handler.db.FindArtistByID(artistID)
	if searchErr != nil {
		if searchErr.(dberror.DBError).ErrorCode != dberror.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to search DB for artist #%v due to: %s\n", artistID, searchErr)
			server.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("Can't find artist #%v\n", artistID)
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found artist #%v: %v\n", artistID, artist.Name)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = convert.WriteToJSON(&artist, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

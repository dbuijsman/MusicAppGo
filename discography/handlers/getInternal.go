package handlers

import (
	"general"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// FindSongByID returns the song that belongs to the given ID.
func (handler *MusicHandler) FindSongByID(response http.ResponseWriter, request *http.Request) {
	songIDstring := mux.Vars(request)["id"]
	songID, err := strconv.Atoi(songIDstring)
	if err != nil {
		handler.Logger.Printf("Received request with invalid id %v results in: %s\n", songIDstring, err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	song, searchErr := handler.db.FindSongByID(songID)
	if searchErr != nil {
		if searchErr.(general.DBError).ErrorCode != general.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to search DB for song #%v due to: %s\n", songID, searchErr)
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("Can't find song #%v\n", songID)
		general.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found song #%v: %v - %v\n", songID, song.Artists, song.Name)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = general.WriteToJSON(&song, response)
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
		general.SendError(response, http.StatusBadRequest)
		return
	}
	artist, searchErr := handler.db.FindArtistByID(artistID)
	if searchErr != nil {
		if searchErr.(general.DBError).ErrorCode != general.NotFoundError {
			handler.Logger.Printf("[ERROR] Failed to search DB for artist #%v due to: %s\n", artistID, searchErr)
			general.SendError(response, http.StatusInternalServerError)
			return
		}
		handler.Logger.Printf("Can't find artist #%v\n", artistID)
		general.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found artist #%v: %v\n", artistID, artist.Name)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = general.WriteToJSON(&artist, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

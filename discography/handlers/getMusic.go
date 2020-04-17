package handlers

import (
	"MusicAppGo/common"
	"discography/database"
	"net/http"

	"github.com/gorilla/mux"
)

// ArtistStartingWith searches the database for artists that satisfies the criria
func (handler *MusicHandler) ArtistStartingWith(response http.ResponseWriter, request *http.Request) {
	firstLetter := mux.Vars(request)["firstLetter"]
	if firstLetter == "0-9" {
		handler.Logger.Printf("[Error] Trying to request non-implemented case %v\n", firstLetter)
		http.Error(response, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
	offset, max, err := common.GetOffsetMaxFromRequest(request)
	if err != nil {
		badRequests.Inc()
		handler.Logger.Printf("%s\n", err)
		http.Error(response, "Invalid query value.", http.StatusBadRequest)
		return
	}
	results, errorSearch := handler.db.GetArtistsStartingWith(firstLetter, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(common.DBError).ErrorCode
		if errorcode == common.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters")
			http.Error(response, errorSearch.Error(), http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("Error %v: %v\n", errorcode, errorSearch.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find artists starting with %v\n", firstLetter)
		failureSearchRequest.Inc()
		http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found artists starting with %v\n", firstLetter)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = common.ToJSON(&MultipleArtists{Music: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

// SongsFromArtist returns a set of songs from the requested artist
func (handler *MusicHandler) SongsFromArtist(response http.ResponseWriter, request *http.Request) {
	nameArtist := mux.Vars(request)["artist"]
	offset, max, err := common.GetOffsetMaxFromRequest(request)
	if err != nil {
		badRequests.Inc()
		handler.Logger.Printf("%s\n", err)
		http.Error(response, "Invalid query value.", http.StatusBadRequest)
		return
	}
	results, errorSearch := handler.db.GetSongsFromArtist(nameArtist, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(common.DBError).ErrorCode
		if errorcode == common.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters")
			http.Error(response, errorSearch.Error(), http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("Error %v: %v\n", errorcode, errorSearch.Error())
		http.Error(response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find songs of %v\n", nameArtist)
		failureSearchRequest.Inc()
		http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	// We need to combine the results
	multipleSongs := make([]database.SongDB, 0, len(results))
	song := database.NewSongDB(results[0].SongID, results[0].SongName)
	for _, rowArtist := range results {
		if rowArtist.SongID != song.ID {
			multipleSongs = append(multipleSongs, song)
			song = database.NewSongDB(rowArtist.SongID, rowArtist.SongName)
		}
		song.Artists = append(song.Artists, database.NewRowArtistDB(rowArtist.ArtistID, rowArtist.ArtistName, rowArtist.ArtistPrefix))
	}
	multipleSongs = append(multipleSongs, song)

	handler.Logger.Printf("Succesfully found songs from %v\n", nameArtist)
	hasNext := (len(multipleSongs) > max)
	if hasNext {
		multipleSongs = multipleSongs[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = common.ToJSON(&MultipleSongs{Music: multipleSongs, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

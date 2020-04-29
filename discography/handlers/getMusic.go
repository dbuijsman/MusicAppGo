package handlers

import (
	"general"
	"net/http"

	"github.com/gorilla/mux"
)

// ArtistStartingWith searches the database for artists that satisfies the criria
func (handler *MusicHandler) ArtistStartingWith(response http.ResponseWriter, request *http.Request) {
	firstLetter := mux.Vars(request)["firstLetter"]
	offsetMax := request.Context().Value(general.OffsetMax{}).(general.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	handler.Logger.Printf("Received call for start %v and limit %v,%v\n", firstLetter, offset, max)
	var results []general.Artist
	var errorSearch error
	if firstLetter == "0-9" {
		results, errorSearch = handler.db.GetArtistsStartingWithNumber(offset, max+1)
	} else {
		results, errorSearch = handler.db.GetArtistsStartingWithLetter(firstLetter, offset, max+1)
	}
	if errorSearch != nil {
		if errorSearch.(general.DBError).ErrorCode == general.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			general.SendError(response, http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("[Error] Can't find artists starting with %v and limit %v,%v due to: %s\n", firstLetter, offset, max, errorSearch)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find artists starting with %v and limit %v,%v\n", firstLetter, offset, max)
		failureSearchRequest.Inc()
		general.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v artists starting with %v and limit %v,%v\n", len(results), firstLetter, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := general.WriteToJSON(&general.MultipleArtists{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

// SongsFromArtist returns a set of songs from the requested artist
func (handler *MusicHandler) SongsFromArtist(response http.ResponseWriter, request *http.Request) {
	nameArtist := mux.Vars(request)["artist"]
	offsetMax := request.Context().Value(general.OffsetMax{}).(general.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	handler.Logger.Printf("Received call for songs of %v and limit %v,%v\n", nameArtist, offset, max)
	results, errorSearch := handler.db.GetSongsFromArtist(nameArtist, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(general.DBError).ErrorCode
		if errorcode == general.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			general.SendError(response, http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("[Error] Can't find songs of %v and limit %v,%v due to: %s\n", nameArtist, offset, max, errorSearch)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find songs of %v and limit %v,%v\n", nameArtist, offset, max)
		failureSearchRequest.Inc()
		general.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v songs of %v and limit %v,%v\n", len(results), nameArtist, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := general.WriteToJSON(&general.MultipleSongs{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

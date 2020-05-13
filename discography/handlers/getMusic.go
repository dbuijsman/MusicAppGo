package handlers

import (
	"fmt"
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// ArtistStartingWith searches the database for artists that satisfies the criria
func (handler *MusicHandler) ArtistStartingWith(response http.ResponseWriter, request *http.Request) {
	firstLetter := mux.Vars(request)["firstLetter"]
	if firstLetter == "undefined" {
		firstLetter = ""
	}
	offsetMax := request.Context().Value(server.OffsetMax{}).(server.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	handler.Logger.Printf("Received call for start %v and limit %v,%v\n", firstLetter, offset, max)
	var results []types.Artist
	var errorSearch error
	if firstLetter == "0-9" {
		results, errorSearch = handler.db.GetArtistsStartingWithNumber(offset, max+1)
	} else {
		results, errorSearch = handler.db.GetArtistsStartingWithLetter(firstLetter, offset, max+1)
	}
	if errorSearch != nil {
		if errorSearch.(dberror.DBError).ErrorCode == dberror.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			server.SendError(response, http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("[Error] Can't find artists starting with %v and limit %v,%v due to: %s\n", firstLetter, offset, max, errorSearch)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find artists starting with %v and limit %v,%v\n", firstLetter, offset, max)
		failureSearchRequest.Inc()
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v artists starting with %v and limit %v,%v\n", len(results), firstLetter, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := convert.WriteToJSON(&types.MultipleArtists{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

// SongsFromArtist returns a set of songs from the requested artist
func (handler *MusicHandler) SongsFromArtist(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(types.Credentials{})
	nameArtist := mux.Vars(request)["artist"]
	offsetMax := request.Context().Value(server.OffsetMax{}).(server.OffsetMax)
	offset, max := offsetMax.Offset, offsetMax.Max
	preferences := make(map[int]string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(user interface{}) {
		defer wg.Done()
		if user == nil {
			return
		}
		userID := user.(types.Credentials).ID
		resp, err := handler.GETRequest(fmt.Sprintf("%v/intern/preference/%v/%v", addressLikes, userID, nameArtist))
		if err != nil || resp.StatusCode != http.StatusOK {
			handler.Logger.Printf("Failed to obtain preferences of user #%v for artist %v due to: %s\n", userID, nameArtist, err)
			return
		}
		if err = convert.ReadFromJSONNoValidation(&preferences, resp.Body); err != nil {
			handler.Logger.Printf("[ERROR] Failed to deserialze map of songsID and preferences due to: %s\n", err)
		}
		handler.Logger.Printf("Received response for user #%v for artist %v: %v\n", userID, nameArtist, resp.StatusCode)

	}(user)
	handler.Logger.Printf("Received call for songs of %v and limit %v,%v\n", nameArtist, offset, max)
	results, errorSearch := handler.db.GetSongsFromArtist(nameArtist, offset, max+1)
	if errorSearch != nil {
		errorcode := errorSearch.(dberror.DBError).ErrorCode
		if errorcode == dberror.InvalidOffsetMax {
			badRequests.Inc()
			handler.Logger.Printf("Request with invalid  values for query parameters: %v,%v", offset, max)
			server.SendError(response, http.StatusBadRequest)
			return
		}
		failureSearchRequest.Inc()
		handler.Logger.Printf("[Error] Can't find songs of %v and limit %v,%v due to: %s\n", nameArtist, offset, max, errorSearch)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	if len(results) == 0 {
		handler.Logger.Printf("Failed to find songs of %v and limit %v,%v\n", nameArtist, offset, max)
		failureSearchRequest.Inc()
		server.SendError(response, http.StatusNotFound)
		return
	}
	handler.Logger.Printf("Succesfully found %v songs of %v and limit %v,%v\n", len(results), nameArtist, offset, max)
	hasNext := (len(results) > max)
	if hasNext {
		results = results[0:max]
	}
	wg.Wait()
	if user != nil {
		for index, song := range results {
			results[index].Preference = preferences[song.ID]
		}
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := convert.WriteToJSON(&types.MultipleSongs{Data: results, HasNext: hasNext}, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

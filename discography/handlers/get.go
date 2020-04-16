package handlers

import (
	"MusicAppGo/common"
	"net/http"

	"github.com/gorilla/mux"
)

// ArtistStartingWith searches the database for artists that satisfies the criria
func (handler *MusicHandler) ArtistStartingWith(response http.ResponseWriter, request *http.Request) {
	firstLetter := mux.Vars(request)["firstLetter"]
	if firstLetter == "" {
		badRequests.Inc()
		handler.Logger.Printf("Got request with no first letter\n")
		http.Error(response, "Bad request.", http.StatusBadRequest)
		return
	}
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
		handler.Logger.Printf("Error %v: %v\n", errorcode, errorSearch.Error())
		http.Error(response, "Internal server error", http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Succesfully found artists starting with %v\n", firstLetter)
	succesNewArtist.Inc()
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

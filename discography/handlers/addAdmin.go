package handlers

import (
	"MusicAppGo/common"
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
	prefix, artist := seperatePrefix(newArtist.Name)
	if err := handler.db.AddArtist(artist, prefix, newArtist.LinkSpotify); err != nil {
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

func seperatePrefix(name string) (prefix, artist string) {
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

// function seperatePrefix(artist){
//     arrayPrefixes = ["A ", "An ", "The "];
//     for(const prefix of arrayPrefixes){
//         if(artist.startsWith(prefix)){
//             artist = artist.substring(prefix.length);
//             return {"prefix": prefix, "name_artist": artist}
//         }
//     }
//     return {"prefix": null, "name_artist": artist}
// }

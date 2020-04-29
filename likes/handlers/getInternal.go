package handlers

import (
	"general"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// GetPreferencesOfArtist responds with a map combining song IDs and the preference (like or dislike)
func (handler *LikesHandler) GetPreferencesOfArtist(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(general.Credentials{}).(general.Credentials)
	nameArtist := mux.Vars(request)["artist"]
	handler.Logger.Printf("Received internal call for preferences of artist %v for user #%v\n", nameArtist, user.ID)
	results := make(map[int]string)
	likesChan := make(chan int, 20)
	dislikesChan := make(chan int, 20)
	doneChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		doneChan <- true
	}(&wg)
	go handler.db.GetLikesIDFromArtistName(handler.Logger, user.ID, nameArtist, likesChan, &wg)
	go handler.db.GetDislikesIDFromArtistName(handler.Logger, user.ID, nameArtist, dislikesChan, &wg)
	for {
		select {
		case like := <-likesChan:
			results[like] = "like"
		case dislike := <-dislikesChan:
			results[dislike] = "dislike"
		case <-doneChan:
			break
		}
	}
	handler.Logger.Printf("User #%v has %v preferences of songs of artist %v\n", user.ID, len(results), nameArtist)
	if len(results) == 0 {
		general.SendError(response, http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err := general.WriteToJSON(&results, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

package handlers

import (
	"general"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// GetPreferencesOfArtist responds with a map combining song IDs and the preference (like or dislike)
func (handler *LikesHandler) GetPreferencesOfArtist(response http.ResponseWriter, request *http.Request) {
	nameArtist := mux.Vars(request)["artist"]
	user := mux.Vars(request)["user"]
	userID, err := strconv.Atoi(user)
	if err != nil {
		general.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received internal call for preferences of artist %v for user #%v\n", nameArtist, userID)
	results := make(map[int]string)
	likesChan := make(chan int, 20)
	dislikesChan := make(chan int, 20)
	doneChan := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		handler.Logger.Printf("Found all preferences of user #%v\n", userID)
		doneChan <- true
	}(&wg)
	go handler.db.GetLikesIDFromArtistName(handler.Logger, userID, nameArtist, likesChan, &wg)
	go handler.db.GetDislikesIDFromArtistName(handler.Logger, userID, nameArtist, dislikesChan, &wg)
LOOP:
	for {
		select {
		case like := <-likesChan:
			results[like] = "like"
		case dislike := <-dislikesChan:
			results[dislike] = "dislike"
		case <-doneChan:
			handler.Logger.Printf("Stop waiting for more results for user #%v\n", userID)
			break LOOP
		}
	}
	handler.Logger.Printf("User #%v has %v preferences of songs of artist %v\n", userID, len(results), nameArtist)
	if len(results) == 0 {
		general.SendError(response, http.StatusNotFound)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = general.WriteToJSON(&results, response)
	if err != nil {
		handler.Logger.Printf("[ERROR] %s\n", err)
	}
}

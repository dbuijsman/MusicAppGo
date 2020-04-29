package handlers

import (
	"general"
	"net/http"
)

// RemoveLike removes a like from the database.
func (handler *LikesHandler) RemoveLike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(general.Credentials{}).(general.Credentials)
	var preference general.Preference
	if err := general.ReadFromJSON(&preference, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for removing a like of user #%v and song #%v\n", user.ID, preference.ID)
	err := handler.db.RemoveLike(user.ID, preference.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to remove like of user #%v and song #%v: %s\n", user.ID, preference.ID, err)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

// RemoveDislike removes a dislike from the database.
func (handler *LikesHandler) RemoveDislike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(general.Credentials{}).(general.Credentials)
	var preference general.Preference
	if err := general.ReadFromJSON(&preference, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for removing a dislike of user #%v and song #%v\n", user.ID, preference.ID)
	err := handler.db.RemoveDislike(user.ID, preference.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to remove dislike of user #%v and song #%v: %s\n", user.ID, preference.ID, err)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

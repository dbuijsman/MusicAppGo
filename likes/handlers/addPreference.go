package handlers

import (
	"bytes"
	"fmt"
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"
	"net/http"
)

func (handler *LikesHandler) obtainSongOrSendError(response http.ResponseWriter, songID int) bool {
	resp, err := handler.GETRequest(fmt.Sprintf("%v/intern/song/%v", addressDiscography, songID))
	if resp.StatusCode == http.StatusNotFound {
		handler.Logger.Printf("Song #%v doesn't exist!\n", songID)
		server.SendError(response, http.StatusNotFound)
		return true
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		handler.Logger.Printf("Failed to obtain song #%v from discography: %s\n", songID, err)
		server.SendError(response, http.StatusInternalServerError)
		return true
	}
	handler.Logger.Printf("Found missing song #%v from discography service\n", songID)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	handler.ConsumeNewSong(buf.Bytes())
	return false
}

// AddLike adds a new like to the database and it deletes a potential dislike to the same song.
func (handler *LikesHandler) AddLike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(types.Credentials{}).(types.Credentials)
	var newPref types.Preference
	if err := convert.ReadFromJSON(&newPref, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		server.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for new like of user #%v and song #%v\n", user.ID, newPref.ID)
	go func(userID, songID int) {
		err := handler.db.RemoveDislike(userID, songID)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to remove dislike of user #%v and song #%v: %s\n", userID, songID, err)
		} else {
			handler.Logger.Printf("Succesfully removed a potential like and dislike situation for user #%v and song #%v\n", userID, songID)
		}
	}(user.ID, newPref.ID)
	err := handler.db.AddLike(user.ID, newPref.ID)
	// If no error occurs, then we can response with StatusOK
	if err == nil {
		handler.Logger.Printf("Succesfully added like for user #%v and song #%v\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	if err.(dberror.DBError).ErrorCode == dberror.DuplicateEntry {
		handler.Logger.Printf("Like for user#%v and song #%v already exists\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	// If err gives an unexpected error, then we will send internal server error
	if err.(dberror.DBError).ErrorCode != dberror.MissingForeignKey {
		handler.Logger.Printf("[ERROR] Failed to add new like for user #%v and song #%v: %s\n", user.ID, newPref.ID, err)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Can't add like for user #%v and song #%v. Trying to add user and song\n", user.ID, newPref.ID)
	channelAddUser := make(chan error)
	go func() {
		channelAddUser <- handler.db.AddUser(user)
	}()
	if handler.obtainSongOrSendError(response, newPref.ID) {
		return
	}
	errAddUser := <-channelAddUser
	if errAddUser != nil && errAddUser.(dberror.DBError).ErrorCode != dberror.DuplicateEntry {
		handler.Logger.Printf("[ERROR] Failed to add new user %v: %s\n", user.Username, errAddUser)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	err = handler.db.AddLike(user.ID, newPref.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to add new like for user #%v and song #%v after adding missing data: %s\n", user.ID, newPref.ID, err)
		server.SendError(response, http.StatusInternalServerError)
		return

	}
	handler.Logger.Printf("Succesfully added like for user #%v and song #%v after adding missing data\n", user.ID, newPref.ID)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

// AddDislike adds a new dislike to the database and it deletes a potential like to the same song.
func (handler *LikesHandler) AddDislike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(types.Credentials{}).(types.Credentials)
	var newPref types.Preference
	if err := convert.ReadFromJSON(&newPref, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		server.SendError(response, http.StatusBadRequest)
		return
	}
	handler.Logger.Printf("Received call for new dislike of user #%v and song #%v\n", user.ID, newPref.ID)
	go func(userID, songID int) {
		err := handler.db.RemoveLike(userID, songID)
		if err != nil {
			handler.Logger.Printf("[ERROR] Failed to remove like of user #%v and song #%v: %s\n", userID, songID, err)
		} else {
			handler.Logger.Printf("Succesfully removed a potential like and dislike situation for user #%v and song #%v\n", userID, songID)
		}
	}(user.ID, newPref.ID)
	err := handler.db.AddDislike(user.ID, newPref.ID)
	// If no error occurs, then we can response with StatusOK
	if err == nil {
		handler.Logger.Printf("Succesfully added dislike for user #%v and song #%v\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	if err.(dberror.DBError).ErrorCode == dberror.DuplicateEntry {
		handler.Logger.Printf("Dislike for user#%v and song #%v already exists\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	// If err gives an unexpected error, then we will send internal server error
	if err.(dberror.DBError).ErrorCode != dberror.MissingForeignKey {
		handler.Logger.Printf("[ERROR] Failed to add new dislike for user #%v and song #%v: %s\n", user.ID, newPref.ID, err)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	handler.Logger.Printf("Can't add dislike for user #%v and song #%v. Trying to add user and song\n", user.ID, newPref.ID)
	channelAddUser := make(chan error)
	go func() {
		channelAddUser <- handler.db.AddUser(user)
	}()
	if handler.obtainSongOrSendError(response, newPref.ID) {
		return
	}
	errAddUser := <-channelAddUser
	if errAddUser != nil && errAddUser.(dberror.DBError).ErrorCode != dberror.DuplicateEntry {
		handler.Logger.Printf("[ERROR] Failed to add new user %v: %s\n", user.Username, errAddUser)
		server.SendError(response, http.StatusInternalServerError)
		return
	}
	err = handler.db.AddDislike(user.ID, newPref.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to add new dislike for user #%v and song #%v after adding missing data: %s\n", user.ID, newPref.ID, err)
		server.SendError(response, http.StatusInternalServerError)
		return

	}
	handler.Logger.Printf("Succesfully added dislike for user #%v and song #%v after adding missing data\n", user.ID, newPref.ID)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

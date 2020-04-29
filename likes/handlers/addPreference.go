package handlers

import (
	"general"
	"net/http"
)

// AddLike adds a new like to the database and it deletes a potential dislike to the same song.
func (handler *LikesHandler) AddLike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(general.Credentials{}).(general.Credentials)
	var newPref general.Preference
	if err := general.ReadFromJSON(&newPref, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
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
	if err.(general.DBError).ErrorCode == general.DuplicateEntry {
		handler.Logger.Printf("Like for user#%v and song #%v already exists\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	// If err gives an unexpected error, then we will send internal server error
	if err.(general.DBError).ErrorCode != general.MissingForeignKey {
		handler.Logger.Printf("[ERROR] Failed to add new like for user #%v and song #%v: %s\n", user.ID, newPref.ID, err)
		general.SendError(response, http.StatusInternalServerError)
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
	if errAddUser != nil && errAddUser.(general.DBError).ErrorCode != general.DuplicateEntry {
		handler.Logger.Printf("[ERROR] Failed to add new user %v: %s\n", user.Username, errAddUser)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	err = handler.db.AddLike(user.ID, newPref.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to add new like for user #%v and song #%v after adding missing data: %s\n", user.ID, newPref.ID, err)
		general.SendError(response, http.StatusInternalServerError)
		return

	}
	handler.Logger.Printf("Succesfully added like for user #%v and song #%v after adding missing data\n", user.ID, newPref.ID)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

// AddDislike adds a new dislike to the database and it deletes a potential like to the same song.
func (handler *LikesHandler) AddDislike(response http.ResponseWriter, request *http.Request) {
	user := request.Context().Value(general.Credentials{}).(general.Credentials)
	var newPref general.Preference
	if err := general.ReadFromJSON(&newPref, request.Body); err != nil {
		badRequests.Inc()
		handler.Logger.Printf("Got invalid request: %v\n", err)
		general.SendError(response, http.StatusBadRequest)
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
	if err.(general.DBError).ErrorCode == general.DuplicateEntry {
		handler.Logger.Printf("Dislike for user#%v and song #%v already exists\n", user.ID, newPref.ID)
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(http.StatusText(http.StatusOK)))
		return
	}
	// If err gives an unexpected error, then we will send internal server error
	if err.(general.DBError).ErrorCode != general.MissingForeignKey {
		handler.Logger.Printf("[ERROR] Failed to add new dislike for user #%v and song #%v: %s\n", user.ID, newPref.ID, err)
		general.SendError(response, http.StatusInternalServerError)
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
	if errAddUser != nil && errAddUser.(general.DBError).ErrorCode != general.DuplicateEntry {
		handler.Logger.Printf("[ERROR] Failed to add new user %v: %s\n", user.Username, errAddUser)
		general.SendError(response, http.StatusInternalServerError)
		return
	}
	err = handler.db.AddDislike(user.ID, newPref.ID)
	if err != nil {
		handler.Logger.Printf("[ERROR] Failed to add new dislike for user #%v and song #%v after adding missing data: %s\n", user.ID, newPref.ID, err)
		general.SendError(response, http.StatusInternalServerError)
		return

	}
	handler.Logger.Printf("Succesfully added dislike for user #%v and song #%v after adding missing data\n", user.ID, newPref.ID)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(http.StatusText(http.StatusOK)))
}

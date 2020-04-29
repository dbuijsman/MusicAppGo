package handlers

import (
	"bytes"
	"fmt"
	"general"
	"net/http"
)

func (handler *LikesHandler) obtainSongOrSendError(response http.ResponseWriter, songID int) bool {
	resp, err := handler.GETRequest(fmt.Sprintf("localhost%v/song/%v", portDiscography, songID))
	if resp.StatusCode == http.StatusNotFound {
		handler.Logger.Printf("Song #%v doesn't exist!\n", songID)
		general.SendError(response, http.StatusNotFound)
		return true
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		handler.Logger.Printf("Failed to obtain song #%v from discography: %s\n", songID, err)
		general.SendError(response, http.StatusInternalServerError)
		return true
	}
	handler.Logger.Printf("Found missing song #%v from discography service\n", songID)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	handler.ConsumeNewSong(buf.Bytes())
	return false
}

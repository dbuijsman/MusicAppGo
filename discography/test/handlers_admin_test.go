package test

import (
	"MusicAppGo/common"
	"discography/database"
	"discography/handlers"
	"log"
	"net/http"
	"testing"
)

func TestAddAdmin_savingInDB(t *testing.T) {
	cases := map[string]struct {
		nameArtist                   string
		expectedPrefix, expectedName string
	}{
		"Artist with a prefix":                                   {"The Rolling Stones", "The", "Rolling Stones"},
		"Artist with no prefix":                                  {"Sum 41", "", "Sum 41"},
		"Artist without prefix but the name starts with one (A)": {"Avenged Sevenfold", "", "Avenged Sevenfold"},
	}
	for nameCase, newArtist := range cases {
		l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
		db := testDB{db: make(map[string]database.RowArtistDB)}
		handler := handlers.NewMusicHandler(l, db).AddArtist
		common.TestPostRequest(t, handler, handlers.NewArtist{Name: newArtist.nameArtist})
		result := db.db[newArtist.expectedName]
		if result.Artist != newArtist.expectedName {
			t.Errorf("Admin adding new artist: %v expects name: %v but got: %v\n", newArtist.nameArtist, newArtist.expectedName, result.Artist)
		}
		if result.Prefix != newArtist.expectedPrefix {
			t.Errorf("%v: Admin adding new artist: %v expects prefix: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expectedPrefix, result.Prefix)
		}
	}
}
func TestAddAdmin_statusCode(t *testing.T) {
	cases := map[string]struct {
		nameArtist string
		expected   int
	}{
		"Artist with a prefix":                                   {"The Rolling Stones", http.StatusOK},
		"Artist with no prefix":                                  {"Sum 41", http.StatusOK},
		"Artist without prefix but the name starts with one (A)": {"Avenged Sevenfold", http.StatusOK},
		"Empty artist name":                                      {"", http.StatusBadRequest},
	}
	for nameCase, newArtist := range cases {
		handler := testMusicHandler()
		response := common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist.nameArtist})
		if response.Code != newArtist.expected {
			t.Errorf("%v: Admin adding new artist: %v expects statuscode: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expected, response.Code)
		}
	}
}
func TestSignUp_duplicateEntry(t *testing.T) {
	someArtist := handlers.NewArtist{Name: "Test", LinkSpotify: "Testing"}
	cases := map[string]struct {
		name, link string
		expected   int
	}{
		"Duplicate input":  {someArtist.Name, someArtist.LinkSpotify, http.StatusUnprocessableEntity},
		"Different input ": {someArtist.Name + "NOT", someArtist.LinkSpotify + "NOT", http.StatusOK},
		"Duplicate name":   {someArtist.Name, someArtist.LinkSpotify + "NOT", http.StatusUnprocessableEntity},
	}
	for nameCase, newArtist := range cases {
		handler := testMusicHandler()
		common.TestPostRequest(t, handler.AddArtist, someArtist)
		recorder := common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist.name, LinkSpotify: newArtist.link})
		if recorder.Code != newArtist.expected {
			t.Errorf("%v: Admin adding new artist: %v and linl: %v after base case expects statuscode: %v but got: %v\n", nameCase, newArtist.name, newArtist.link, newArtist.expected, recorder.Code)
		}
	}
}

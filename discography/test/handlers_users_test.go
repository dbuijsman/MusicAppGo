package test

import (
	"MusicAppGo/common"
	"discography/handlers"
	"fmt"
	"net/http"
	"testing"
)

func TestArtistStartingWith_statusCode(t *testing.T) {
	cases := map[string]struct {
		offset, max        string
		expectedStatusCode int
	}{
		"Negative offset":        {"-1", "10", http.StatusBadRequest},
		"Negative max":           {"0", "-1", http.StatusBadRequest},
		"Correct offset and max": {"1", "10", http.StatusOK},
		"Default offset and max": {"0", "20", http.StatusOK},
		"Offset is NaN":          {"text", "10", http.StatusBadRequest},
		"Max is NaN":             {"0", "text", http.StatusBadRequest},
	}
	for nameCase, queryValues := range cases {
		handler := testMusicHandler()
		common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: "Disturbed"})
		common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: "Dio"})
		query := fmt.Sprintf("offset=%v&max=%v", queryValues.offset, queryValues.max)
		response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", "D", query)
		if response.Code != queryValues.expectedStatusCode {
			t.Errorf("%v: Expects statuscode %v but got %v\n", nameCase, queryValues.expectedStatusCode, response.Code)
		}
	}
}
func TestArtistStartingWith_amountResults(t *testing.T) {
	cases := map[string]struct {
		artists               []string
		firstLetter           string
		offset, max           int
		expectedAmountResults int
	}{
		"Artist without prefix with right first letter":         {[]string{"Bob Dylan"}, "B", 0, 10, 1},
		"Artist with prefix with right first letter":            {[]string{"The Beatles"}, "B", 0, 10, 1},
		"Prefix does not count as a first letter":               {[]string{"The Beatles"}, "T", 0, 10, 0},
		"Only the first letter counts":                          {[]string{"Bob Dylan"}, "D", 0, 10, 0},
		"Multiple artists with right first letter":              {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 0, 10, 3},
		"Amount of results capped by max":                       {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 0, 2, 2},
		"Skip amount of results given by offset":                {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 2, 10, 1},
		"Correctly handle combination of max and offset":        {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 1, 2, 2},
		"Return only the artists with the correct first letter": {[]string{"Pendulum", "The Prodigy", "Bob Dylan"}, "P", 0, 10, 2},
	}
	for nameCase, discography := range cases {
		handler := testMusicHandler()
		for _, newArtist := range discography.artists {
			common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist})
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query)
		var result handlers.MultipleArtists
		err := common.FromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if len(result.Music) != discography.expectedAmountResults {
			t.Errorf("%v: Searching for artists starting with %v, offset %v and max %v should give %v results but got %v\n", nameCase, discography.firstLetter, discography.offset, discography.max, discography.expectedAmountResults, len(result.Music))
		}
	}
}

func TestArtistStartingWith_correctOrderResults(t *testing.T) {
	cases := map[string]struct {
		artists           []string
		firstLetter       string
		offset, max       int
		indexExpectedName int
		expectedName      string
	}{
		"Order the result":                          {[]string{"The Beatles", "Bon Jovi", "Bob Dylan"}, "B", 0, 10, 2, "Bon Jovi"},
		"Skip artists when offset is given":         {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 2, 10, 0, "Bob Dylan"},
		"Return only results before max is reached": {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 0, 3, 2, "Bob Dylan"},
	}
	for nameCase, discography := range cases {
		handler := testMusicHandler()
		for _, newArtist := range discography.artists {
			common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist})
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query)
		var result handlers.MultipleArtists
		err := common.FromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] Decoding response: %v\n", err)
		}
		if artist := result.Music[discography.indexExpectedName].Artist; artist != discography.expectedName {
			t.Errorf("%v: Searching for artists with given first letter expects %v at index %v but got %v\n", nameCase, discography.expectedName, discography.indexExpectedName, artist)
		}
	}
}
func TestArtistStartingWith_missingOffset(t *testing.T) {
	handler := testMusicHandler()
	common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: "Disturbed"})
	response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", "D", "max=10")
	var result handlers.MultipleArtists
	err := common.FromJSON(&result, response.Body)
	if err != nil {
		t.Fatalf("[ERROR] Decoding response: %v\n", err)
	}
	if len(result.Music) != 1 {
		t.Errorf("Sendig request without offset to first letter should return results")
	}
}
func TestArtistStartingWith_missingMax(t *testing.T) {
	handler := testMusicHandler()
	common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: "Disturbed"})
	response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", "D", "offset=0")
	var result handlers.MultipleArtists
	err := common.FromJSON(&result, response.Body)
	if err != nil {
		t.Fatalf("[ERROR] Decoding response: %v\n", err)
	}
	if len(result.Music) != 1 {
		t.Errorf("Sendig request without max to first letter should return results")
	}
}
func TestArtistStartingWith_HasNextResult(t *testing.T) {
	cases := map[string]struct {
		artists         []string
		firstLetter     string
		offset, max     int
		expectedHasNext bool
	}{
		"Max bigger than amount of results":           {[]string{"The Beatles", "Bon Jovi", "Bob Dylan"}, "B", 0, 10, false},
		"Offset + max smaller than amount of results": {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 0, 2, true},
		"Offset + max bigger than amount of results":  {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 3, 3, false},
		"Offset + max equal to amount of results":     {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 2, 2, false},
	}
	for nameCase, discography := range cases {
		handler := testMusicHandler()
		for _, newArtist := range discography.artists {
			common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist})
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := common.TestGetRequestWithPathVariable(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query)
		var result handlers.MultipleArtists
		err := common.FromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] Decoding response: %v\n", err)
		}
		if hasNext := result.HasNext; hasNext != discography.expectedHasNext {
			t.Errorf("%v: Expects hasNext equal to: %v but got: %v\n", nameCase, discography.expectedHasNext, hasNext)
		}
	}
}

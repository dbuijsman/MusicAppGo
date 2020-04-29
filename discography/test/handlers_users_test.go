package test

import (
	"discography/handlers"
	"fmt"
	"general"
	"net/http"
	"testing"
)

func TestArtistStartingWith_statusCode(t *testing.T) {
	cases := map[string]struct {
		artists            []string
		firstLetter        string
		offset             int
		expectedStatusCode int
	}{
		"Artist is found":                         {[]string{"Dio", "Disturbed"}, "D", 0, http.StatusOK},
		"Artist starts with a number":             {[]string{"30 Seconds To Mars", "50 Cent"}, "0-9", 0, http.StatusOK},
		"No artist is found":                      {[]string{"Dio", "Disturbed"}, "Q", 0, http.StatusNotFound},
		"Offset is bigger than amount of results": {[]string{"Dio", "Disturbed"}, "D", 2, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		for _, nameArtist := range testCase.artists {
			general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(nameArtist, ""))
		}
		query := fmt.Sprintf("offset=%v", testCase.offset)
		response := general.TestGetRequestWithPath(t, handler.ArtistStartingWith, "firstLetter", testCase.firstLetter, query, general.GetOffsetMaxMiddleware)
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode %v but got %v\n", nameCase, testCase.expectedStatusCode, response.Code)
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
		"Artist without prefix with right first letter":  {[]string{"Bob Dylan"}, "B", 0, 10, 1},
		"Artist with prefix with right first letter":     {[]string{"The Beatles"}, "B", 0, 10, 1},
		"Prefix does not count as a first letter":        {[]string{"The Beatles", "Tenacious D"}, "T", 0, 10, 1},
		"Only the first letter counts":                   {[]string{"Bob Dylan", "D12"}, "D", 0, 10, 1},
		"Multiple artists with right first letter":       {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 0, 10, 3},
		"Amount of results capped by max":                {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 0, 2, 2},
		"Skip amount of results given by offset":         {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 2, 10, 1},
		"Correctly handle combination of max and offset": {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 1, 2, 2},
		"Only artists with the correct first letter":     {[]string{"Pendulum", "The Prodigy", "Bob Dylan"}, "P", 0, 10, 2},
		"Artist starting with a number":                  {[]string{"50 Cent"}, "0-9", 0, 10, 1},
		"Multiple artist starting with a number":         {[]string{"10CC", "50 Cent"}, "0-9", 0, 10, 2},
		"Only artists starting with a number":            {[]string{"10CC", "50 Cent", "Avenged Sevenfold"}, "0-9", 0, 10, 2},
		"Starting with number only looks at start":       {[]string{"10CC", "50 Cent", "Tech N9ne"}, "0-9", 0, 10, 2},
	}
	for nameCase, discography := range cases {
		handler := testMusicHandler()
		for _, newArtist := range discography.artists {
			general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist, ""))
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := general.TestGetRequestWithPath(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleArtists
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if len(result.Data) != discography.expectedAmountResults {
			t.Errorf("%v: Searching for artists starting with %v, offset %v and max %v should give %v results but got %v\n", nameCase, discography.firstLetter, discography.offset, discography.max, discography.expectedAmountResults, len(result.Data))
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
		"Order the result":                          {[]string{"The Beatles", "Bon Jovi", "Bob Dylan"}, "B", 0, 10, 1, "Bob Dylan"},
		"Correct first result":                      {[]string{"The Beatles", "Bon Jovi", "Bob Dylan"}, "B", 0, 10, 0, "Beatles"},
		"Correct last result":                       {[]string{"The Beatles", "Bon Jovi", "Bob Dylan"}, "B", 0, 10, 2, "Bon Jovi"},
		"Skip artists when offset is given":         {[]string{"The Beatles", "The Bee Gees", "Bob Dylan"}, "B", 2, 10, 0, "Bob Dylan"},
		"Return only results before max is reached": {[]string{"The Beatles", "The Bee Gees", "Bob Dylan", "Bon Jovi"}, "B", 0, 3, 2, "Bob Dylan"},
	}
	for nameCase, discography := range cases {
		handler := testMusicHandler()
		for _, newArtist := range discography.artists {
			general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist, ""))
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := general.TestGetRequestWithPath(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleArtists
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if artist := result.Data[discography.indexExpectedName].Name; artist != discography.expectedName {
			t.Errorf("%v: Searching for artists with given first letter expects %v at index %v but got %v\n", nameCase, discography.expectedName, discography.indexExpectedName, artist)
		}
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
			general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist, ""))
		}
		query := fmt.Sprintf("offset=%v&max=%v", discography.offset, discography.max)
		response := general.TestGetRequestWithPath(t, handler.ArtistStartingWith, "firstLetter", discography.firstLetter, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleArtists
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if hasNext := result.HasNext; hasNext != discography.expectedHasNext {
			t.Errorf("%v: Expects hasNext equal to: %v but got: %v\n", nameCase, discography.expectedHasNext, hasNext)
		}
	}
}

func TestSongsFromArtist_statusCode(t *testing.T) {
	cases := map[string]struct {
		existingSong       song
		artist             string
		offset             int
		expectedStatusCode int
	}{
		"Artist has some songs":                        {getNewSong("The Sound of Silence", "Disturbed"), "Disturbed", 0, http.StatusOK},
		"No song is found":                             {getNewSong("The Sound of Silence", "Disturbed"), "Dio", 0, http.StatusNotFound},
		"Offset is bigger than total amount of result": {getNewSong("The Sound of Silence", "Disturbed"), "Disturbed", 1, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		handler.AddSong(testCase.existingSong.nameSong, testCase.existingSong.artists...)
		query := fmt.Sprintf("offset=%v", testCase.offset)
		response := general.TestGetRequestWithPath(t, handler.SongsFromArtist, "artist", testCase.artist, query, general.GetOffsetMaxMiddleware)
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", nameCase, testCase.expectedStatusCode, response.Code)
		}
	}
}
func TestSongsFromArtist_amountResults(t *testing.T) {
	cases := map[string]struct {
		songs                 []song
		artist                string
		offset, max           int
		expectedAmountResults int
	}{
		"Only songs of the requested artist":             {[]song{getNewSong("The Sound of Silence", "Disturbed"), getNewSong("Shout 2000", "Disturbed"), getNewSong("Last Resort", "Papa Roach")}, "Disturbed", 0, 10, 2},
		"Includes collaborations":                        {[]song{getNewSong("Crazy", "Lost Frequencies", "Zonderling"), getNewSong("Reality", "Lost Frequencies")}, "Lost Frequencies", 0, 10, 2},
		"Amount of results capped by max":                {[]song{getNewSong("My Band", "D12"), getNewSong("Fight Music", "D12"), getNewSong("Fame", "D12")}, "D12", 0, 2, 2},
		"Skip amount of results given by offset":         {[]song{getNewSong("My Band", "D12"), getNewSong("Fight Music", "D12"), getNewSong("Fame", "D12")}, "D12", 2, 10, 1},
		"Correctly handle combination of max and offset": {[]song{getNewSong("My Band", "D12"), getNewSong("Fight Music", "D12"), getNewSong("Fame", "D12"), getNewSong("Rap Game", "D12")}, "D12", 1, 2, 2},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		for _, song := range testCase.songs {
			handler.AddSong(song.nameSong, song.artists...)
		}
		query := fmt.Sprintf("offset=%v&max=%v", testCase.offset, testCase.max)
		response := general.TestGetRequestWithPath(t, handler.SongsFromArtist, "artist", testCase.artist, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleSongs
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if len(result.Data) != testCase.expectedAmountResults {
			fmt.Println(result.Data)
			t.Errorf("%v: Searching for songs of %v, offset %v and max %v should give %v results but got %v\n", nameCase, testCase.artist, testCase.offset, testCase.max, testCase.expectedAmountResults, len(result.Data))
		}
	}
}

func TestSongsFromArtist_correctOrderResults(t *testing.T) {
	discography := []song{getNewSong("It's My Life", "Bon Jovi"), getNewSong("Bed of Roses", "Bon Jovi"), getNewSong("Living on a Prayer", "Bon Jovi"), getNewSong("The River", "Bruce Springsteen")}
	cases := map[string]struct {
		artist            string
		offset, max       int
		indexExpectedName int
		expectedName      string
	}{
		"Order the result":                          {"Bon Jovi", 0, 10, 1, "It's My Life"},
		"Correct first result":                      {"Bon Jovi", 0, 10, 0, "Bed of Roses"},
		"Correct last result":                       {"Bon Jovi", 0, 10, 2, "Living on a Prayer"},
		"Skip artists when offset is given":         {"Bon Jovi", 2, 10, 0, "Living on a Prayer"},
		"Return only results before max is reached": {"Bon Jovi", 0, 2, 1, "It's My Life"},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		for _, song := range discography {
			handler.AddSong(song.nameSong, song.artists...)
		}
		query := fmt.Sprintf("offset=%v&max=%v", testCase.offset, testCase.max)
		response := general.TestGetRequestWithPath(t, handler.SongsFromArtist, "artist", testCase.artist, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleSongs
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if song := result.Data[testCase.indexExpectedName].Name; song != testCase.expectedName {
			t.Errorf("%v: Searching for artists with given first letter expects %v at index %v but got %v\n", nameCase, testCase.expectedName, testCase.indexExpectedName, song)
		}
	}
}
func TestSongsFromArtist_HasNextResult(t *testing.T) {
	discography := []song{getNewSong("Stairway to Heaven", "Led Zeppelin"), getNewSong("Immigrant Song", "Led Zeppelin"), getNewSong("Whole Lotta Love", "Led Zeppelin"), getNewSong("Black Dog", "Led Zeppelin")}
	cases := map[string]struct {
		artist          string
		offset, max     int
		expectedHasNext bool
	}{
		"Max bigger than amount of results":           {"Led Zeppelin", 0, 10, false},
		"Offset + max smaller than amount of results": {"Led Zeppelin", 0, 2, true},
		"Offset + max bigger than amount of results":  {"Led Zeppelin", 3, 3, false},
		"Offset + max equal to amount of results":     {"Led Zeppelin", 2, 2, false},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		for _, song := range discography {
			handler.AddSong(song.nameSong, song.artists...)
		}
		query := fmt.Sprintf("offset=%v&max=%v", testCase.offset, testCase.max)
		response := general.TestGetRequestWithPath(t, handler.SongsFromArtist, "artist", testCase.artist, query, general.GetOffsetMaxMiddleware)
		var result general.MultipleSongs
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		if hasNext := result.HasNext; hasNext != testCase.expectedHasNext {
			t.Errorf("%v: Expects hasNext equal to: %v but got: %v\n", nameCase, testCase.expectedHasNext, hasNext)
		}
	}
}
func TestSongsFromArtist_collaborations(t *testing.T) {
	cases := map[string]struct {
		song                        song
		artist                      string
		expectedAmountOfArtists     int
		expectedCollaboratingArtist string
	}{
		"Single artist":              {getNewSong("Black Betty", "Ram Jam"), "Ram Jam", 1, "Ram Jam"},
		"Correct leading artist":     {getNewSong("Crazy", "Lost Frequencies", "Zonderling"), "Lost Frequencies", 2, "Lost Frequencies"},
		"Correct non-leading artist": {getNewSong("Crazy", "Lost Frequencies", "Zonderling"), "Lost Frequencies", 2, "Zonderling"},
	}
	for nameCase, testCase := range cases {
		handler := testMusicHandler()
		handler.AddSong(testCase.song.nameSong, testCase.song.artists...)
		response := general.TestGetRequestWithPath(t, handler.SongsFromArtist, "artist", testCase.artist, "", general.GetOffsetMaxMiddleware)
		var result general.MultipleSongs
		err := general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Fatalf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
		}
		foundArtists := result.Data[0].Artists
		if len(foundArtists) != testCase.expectedAmountOfArtists {
			t.Errorf("%v: Expects %v collaborating artists but got: %v\n", nameCase, testCase.expectedAmountOfArtists, len(foundArtists))
		}
		foundExpectedArtist := false
		for _, artist := range foundArtists {
			if artist.Name == testCase.expectedCollaboratingArtist {
				foundExpectedArtist = true
				break
			}
		}
		if !foundExpectedArtist {
			t.Errorf("%v: Expects to find %v in song %v - %v, but can't find it\n", nameCase, testCase.expectedCollaboratingArtist, testCase.song.artists, testCase.song.nameSong)
		}
	}
}

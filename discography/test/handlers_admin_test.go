package test

import (
	"discography/handlers"
	"general"
	"net/http"
	"sync"
	"testing"
)

func TestAddArtist_savingInDB(t *testing.T) {
	cases := map[string]struct {
		nameArtist                   string
		expectedPrefix, expectedName string
	}{
		"Artist with a prefix":                                   {"The Rolling Stones", "The", "Rolling Stones"},
		"Artist with no prefix":                                  {"Sum 41", "", "Sum 41"},
		"Artist without prefix but the name starts with one (A)": {"Avenged Sevenfold", "", "Avenged Sevenfold"},
	}
	for nameCase, newArtist := range cases {
		db := newTestDB()
		handler := handlers.NewMusicHandler(general.TestEmptyLogger(), db, general.TestSendMessageEmpty())
		general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist.nameArtist, ""))
		result := db.artistsDB[newArtist.expectedName]
		if result.name != newArtist.expectedName {
			t.Errorf("Admin adding new artist: %v expects name: %v but got: %v\n", newArtist.nameArtist, newArtist.expectedName, result.name)
		}
		if result.prefix != newArtist.expectedPrefix {
			t.Errorf("%v: Admin adding new artist: %v expects prefix: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expectedPrefix, result.prefix)
		}
	}
}

func TestAddArtist_statusCode(t *testing.T) {
	cases := map[string]struct {
		nameArtist         string
		expectedStatusCode int
	}{
		"Artist with a prefix":                                   {"The Rolling Stones", http.StatusOK},
		"Artist with no prefix":                                  {"Sum 41", http.StatusOK},
		"Artist without prefix but the name starts with one (A)": {"Avenged Sevenfold", http.StatusOK},
		"Empty artist name":                                      {"", http.StatusBadRequest},
	}
	for nameCase, newArtist := range cases {
		handler := testMusicHandler()
		response := general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist.nameArtist, ""))
		if response.Code != newArtist.expectedStatusCode {
			t.Errorf("%v: Admin adding new artist: %v expects statuscode: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expectedStatusCode, response.Code)
		}
	}
}

func TestAddArtist_duplicateEntry(t *testing.T) {
	someArtist := handlers.NewClientArtist("Test Artist", "Link test")
	cases := map[string]struct {
		name, link         string
		expectedStatusCode int
	}{
		"Duplicate input":  {someArtist.Artist, someArtist.LinkSpotify, http.StatusUnprocessableEntity},
		"Different input ": {someArtist.Artist + "NOT", someArtist.LinkSpotify + "NOT", http.StatusOK},
		"Duplicate name":   {someArtist.Artist, someArtist.LinkSpotify + "NOT", http.StatusUnprocessableEntity},
	}
	for nameCase, newArtist := range cases {
		handler := testMusicHandler()
		general.TestPostRequest(t, handler.AddArtist, someArtist)
		recorder := general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(newArtist.name, newArtist.link))
		if recorder.Code != newArtist.expectedStatusCode {
			t.Errorf("%v: Admin adding new artist: %v and link: %v after base case expects statuscode: %v but got: %v\n", nameCase, newArtist.name, newArtist.link, newArtist.expectedStatusCode, recorder.Code)
		}
	}
}

func TestAddArtist_sendNoMessageWhenRequestFails(t *testing.T) {
	handler := testMusicHandler()
	handler.SendMessage = func(topic string, message []byte) {
		t.Errorf("Adding no new artist expects no new message, but it sends %s to topic %v\n", message, topic)
	}
	general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist("", ""))
}

func TestAddArtist_sendMessage(t *testing.T) {
	cases := map[string]struct {
		artist                               string
		expectedTopic, expectedArtistMessage string
	}{
		"Adding a new artist":             {"Pendulum", "newArtist", "Pendulum"},
		"Adding a new artist with prefix": {"The Offspring", "newArtist", "Offspring"},
	}
	for nameCase, testCase := range cases {
		artist := handlers.NewClientArtist(testCase.artist, "")
		handler := testMusicHandler()
		var wg sync.WaitGroup
		top, msg, sendMessage := general.TestSendMessage(&wg)
		wg.Add(1)
		handler.SendMessage = sendMessage
		general.TestPostRequest(t, handler.AddArtist, artist)
		wg.Wait()
		var result general.Artist
		err := general.FromJSONBytes(&result, []byte(*msg))
		if err != nil {
			t.Errorf("%v: Expects to send a message containing an artist but deserializing results in: %v\n", nameCase, err)
		}
		if *top != testCase.expectedTopic {
			t.Errorf("%v: Expects to send a message to topic %v but instead it was send to %v\n", nameCase, testCase.expectedTopic, *top)
		}
		if result.Name != testCase.expectedArtistMessage {
			t.Errorf("%v: Expects to send artist %v as message but instead it sends: %v\n", nameCase, testCase.expectedArtistMessage, result.Name)
		}
	}
}

func TestAddSong_savingInDB(t *testing.T) {
	cases := map[string]struct {
		existingArtists []string
		artistsNewSong  []string
		song            string
		expectedArtist  string
		expected        bool
	}{
		"New artist":          {nil, []string{"Sum 41"}, "Fatlip", "Sum 41", true},
		"Existing artist":     {[]string{"Sum 41"}, []string{"Sum 41"}, "Still Waiting", "Sum 41", true},
		"Song without artist": {[]string{"Sum 41"}, nil, "Still Waiting", "Sum 41", false},
		"Song of existing artist without name of the song":             {[]string{"Sum 41"}, []string{"Sum 41"}, "", "Sum 41", false},
		"Song of a new artist without the name of the song":            {nil, []string{"Sum 41"}, "", "Sum 41", false},
		"Song with multiple artists gets added to collabaring artists": {nil, []string{"Iggy Pop", "Sum 41"}, "Little Know It All", "Sum 41", true},
		"Song with multiple artists gets added to main artist":         {nil, []string{"Iggy Pop", "Sum 41"}, "Little Know It All", "Iggy Pop", true},
		"New song gets only added to collaborating artists":            {[]string{"The Rolling Stones", "Billy Talent"}, []string{"The Rolling Stones"}, "Sympathy for the Devil", "Billy Talent", false},
	}
	for nameCase, newSong := range cases {
		db := newTestDB()
		handler := handlers.NewMusicHandler(general.TestEmptyLogger(), db, general.TestSendMessageEmpty())
		for _, artist := range newSong.existingArtists {
			general.TestPostRequest(t, handler.AddArtist, handlers.NewClientArtist(artist, ""))
		}
		handler.AddSong(newSong.song, newSong.artistsNewSong...)
		discography := db.songsDB[newSong.expectedArtist]
		if _, okSong := discography[newSong.song]; okSong != newSong.expected {
			t.Errorf("%v: Adding new song expected to be saved in DB: %v but got %v\n", nameCase, newSong.expected, okSong)
		}
	}
}

func TestAddSong_duplicateInput(t *testing.T) {
	someSong := struct{ artist, song string }{"Billy Talent", "Fallen Leaves"}
	cases := map[string]struct {
		artists           []string
		song              string
		expectedErrorCode int
	}{
		"Duplicate input":                                      {[]string{someSong.artist}, someSong.song, general.DuplicateEntry},
		"New song for existing artist":                         {[]string{someSong.artist}, someSong.song + "NOT", 0},
		"New song for different artist but with the same name": {[]string{someSong.artist + "NOT"}, someSong.song, 0},
	}
	for nameCase, newSong := range cases {
		handler := testMusicHandler()
		handler.AddSong(someSong.song, someSong.artist)
		_, err := handler.AddSong(newSong.song, newSong.artists...)
		if newSong.expectedErrorCode == 0 {
			if err != nil {
				t.Errorf("%v: No errorcode but got %v\n", nameCase, err.(general.DBError).ErrorCode)
			}
			continue
		}
		if err == nil {
			t.Errorf("%v: Expected errorcode %v but got no error\n", nameCase, newSong.expectedErrorCode)
			continue
		}
		if errorcode := err.(general.DBError).ErrorCode; errorcode != newSong.expectedErrorCode {
			t.Errorf("%v: Expected errorcode: %v but got: %v\n", nameCase, newSong.expectedErrorCode, errorcode)
		}
	}
}

func TestAddSong_sendMessageNewSong(t *testing.T) {
	artist := "Pendulum"
	newSong := "Slam"
	topic := "newSong"
	handler := testMusicHandler()
	var wg sync.WaitGroup
	msg, sendMessage := general.TestSendMessageToParticularTopic(&wg, topic)
	handler.SendMessage = sendMessage
	wg.Add(1)
	handler.AddSong(newSong, artist)
	wg.Wait()
	var result general.Song
	err := general.FromJSONBytes(&result, []byte(*msg))
	if err != nil {
		t.Errorf("Adding new song: Expects to send a message containing an artist but deserializing results in: %v\n", err)
	}
	if result.Name != newSong {
		t.Errorf("Adding new song: Expects to send artist %v as message but instead it sends: %v\n", newSong, result.Name)
	}
}

func TestAddSong_sendMessageNewArtist(t *testing.T) {
	newArtist := "Pendulum"
	song := "Slam"
	topic := "newArtist"
	handler := testMusicHandler()
	var wg sync.WaitGroup
	msg, sendMessage := general.TestSendMessageToParticularTopic(&wg, topic)
	handler.SendMessage = sendMessage
	wg.Add(1)
	handler.AddSong(song, newArtist)
	wg.Wait()
	var result general.Artist
	err := general.FromJSONBytes(&result, []byte(*msg))
	if err != nil {
		t.Errorf("Adding new song: Expects to send a message containing an artist but deserializing results in: %v\n", err)
	}
	if result.Name != newArtist {
		t.Errorf("Adding new song: Expects to send artist %v as message but instead it sends: %v\n", newArtist, result.Name)
	}
}

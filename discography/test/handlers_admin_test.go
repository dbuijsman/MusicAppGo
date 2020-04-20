package test

import (
	"MusicAppGo/common"
	"discography/database"
	"discography/handlers"
	"log"
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
		l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
		db := newTestDB()
		handler := handlers.NewMusicHandler(l, db, func(topic string, message []byte) error {
			return nil
		})
		common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist.nameArtist})
		result := db.artistsDB[newArtist.expectedName]
		if result.Artist != newArtist.expectedName {
			t.Errorf("Admin adding new artist: %v expects name: %v but got: %v\n", newArtist.nameArtist, newArtist.expectedName, result.Artist)
		}
		if result.Prefix != newArtist.expectedPrefix {
			t.Errorf("%v: Admin adding new artist: %v expects prefix: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expectedPrefix, result.Prefix)
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
		response := common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist.nameArtist})
		if response.Code != newArtist.expectedStatusCode {
			t.Errorf("%v: Admin adding new artist: %v expects statuscode: %v but got: %v\n", nameCase, newArtist.nameArtist, newArtist.expectedStatusCode, response.Code)
		}
	}
}
func TestAddArtist_duplicateEntry(t *testing.T) {
	someArtist := handlers.NewArtist{Name: "Test", LinkSpotify: "Testing"}
	cases := map[string]struct {
		name, link         string
		expectedStatusCode int
	}{
		"Duplicate input":  {someArtist.Name, someArtist.LinkSpotify, http.StatusUnprocessableEntity},
		"Different input ": {someArtist.Name + "NOT", someArtist.LinkSpotify + "NOT", http.StatusOK},
		"Duplicate name":   {someArtist.Name, someArtist.LinkSpotify + "NOT", http.StatusUnprocessableEntity},
	}
	for nameCase, newArtist := range cases {
		handler := testMusicHandler()
		common.TestPostRequest(t, handler.AddArtist, someArtist)
		recorder := common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: newArtist.name, LinkSpotify: newArtist.link})
		if recorder.Code != newArtist.expectedStatusCode {
			t.Errorf("%v: Admin adding new artist: %v and link: %v after base case expects statuscode: %v but got: %v\n", nameCase, newArtist.name, newArtist.link, newArtist.expectedStatusCode, recorder.Code)
		}
	}
}
func TestAddArtist_sendMessage(t *testing.T) {
	topicValue, msgValue := "", ""
	top, msg := &topicValue, &msgValue
	artist := handlers.NewArtist{Name: "Pendulum"}
	expectedTopic := "newArtist"
	expectedMessage, err := common.ToJSONBytes(database.NewRowArtistDB(0, artist.Name, ""))
	if err != nil {
		t.Fatalf("[ERROR] Can't serialize %v to expected message:%s\n", artist.Name, err)
	}
	handler := testMusicHandler()
	var wg sync.WaitGroup
	wg.Add(1)
	handler.SendMessage = func(topic string, message []byte) {
		*top = topic
		*msg = string(message)
		wg.Done()
	}
	common.TestPostRequest(t, handler.AddArtist, artist)
	wg.Wait()
	if *top != expectedTopic {
		t.Errorf("Adding an artist expects to send a message to topic %v but instead it was send to %v\n", expectedTopic, *top)
	}
	if *msg != string(expectedMessage) {
		t.Errorf("Adding an artist expects to send artist %v as message but instead it sends: %v\n", string(expectedMessage), *msg)
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
		l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
		db := newTestDB()
		handler := handlers.NewMusicHandler(l, db, func(topic string, message []byte) error {
			return nil
		})
		for _, artist := range newSong.existingArtists {
			common.TestPostRequest(t, handler.AddArtist, handlers.NewArtist{Name: artist})
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
		"Duplicate input":                                      {[]string{someSong.artist}, someSong.song, common.DuplicateEntry},
		"New song for existing artist":                         {[]string{someSong.artist}, someSong.song + "NOT", 0},
		"New song for different artist but with the same name": {[]string{someSong.artist + "NOT"}, someSong.song, 0},
	}
	for nameCase, newSong := range cases {
		handler := testMusicHandler()
		handler.AddSong(someSong.song, someSong.artist)
		_, err := handler.AddSong(newSong.song, newSong.artists...)
		if newSong.expectedErrorCode == 0 {
			if err != nil {
				t.Errorf("%v: No errorcode but got %v\n", nameCase, err.(common.DBError).ErrorCode)
			}
			continue
		}
		if err == nil {
			t.Errorf("%v: Expected errorcode %v but got no error\n", nameCase, newSong.expectedErrorCode)
			continue
		}
		if errorcode := err.(common.DBError).ErrorCode; errorcode != newSong.expectedErrorCode {
			t.Errorf("%v: Expected errorcode: %v but got: %v\n", nameCase, newSong.expectedErrorCode, errorcode)
		}
	}
}
func TestAddSong_sendMessage(t *testing.T) {
	topicValue, msgValue := "", ""
	top, msg := &topicValue, &msgValue
	song := struct{ artist, song string }{"Billy Talent", "Fallen Leaves"}
	songAsSongDB := database.NewSongDB(0, song.song)
	songAsSongDB.Artists = append(songAsSongDB.Artists, database.NewRowArtistDB(0, song.artist, ""))
	expectedTopic := "newSong"
	expectedMessage, err := common.ToJSONBytes(songAsSongDB)
	if err != nil {
		t.Fatalf("[ERROR] Can't serialize %v - %v to expected message:%s\n", song.artist, song.song, err)
	}
	handler := testMusicHandler()
	var wg sync.WaitGroup
	wg.Add(1)
	handler.SendMessage = func(topic string, message []byte) {
		*top = topic
		*msg = string(message)
		wg.Done()
	}
	handler.AddSong(song.song, song.artist)
	wg.Wait()
	if *top != expectedTopic {
		t.Errorf("Adding an artist expects to send a message to topic %v but instead it was send to %v\n", expectedTopic, *top)
	}
	if *msg != string(expectedMessage) {
		t.Errorf("Adding an artist expects to send artist %v as message but instead it sends: %v\n", string(expectedMessage), *msg)
	}
}

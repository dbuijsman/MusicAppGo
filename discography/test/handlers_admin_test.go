package test

import (
	"discography/handlers"
	"general"
	"net/http"
	"testing"
	"time"
)

func TestAdminHandlers_response(t *testing.T) {
	artist := general.NewArtist(1, "Prodigy", "The")
	song := "Firestarter"
	linkArtist := "linkToArtist"
	cases := map[string]struct {
		path               string
		roleClient         string
		body               interface{}
		expectedStatusCode int
	}{
		"AddArtist: Request without body":                     {"/admin/artist", "admin", nil, http.StatusBadRequest},
		"AddArtist: Artist with a prefix":                     {"/admin/artist", "admin", handlers.NewClientArtist("The Rolling Stones", "link"), http.StatusOK},
		"AddArtist: Artist with no prefix":                    {"/admin/artist", "admin", handlers.NewClientArtist("Sum 41", "link"), http.StatusOK},
		"AddArtist: Artist starting with A (no prefix)":       {"/admin/artist", "admin", handlers.NewClientArtist("Avenged Sevenfold", "link"), http.StatusOK},
		"AddArtist: Empty artist name":                        {"/admin/artist", "admin", handlers.NewClientArtist("", "link"), http.StatusBadRequest},
		"AddArtist: No Spotify link":                          {"/admin/artist", "admin", handlers.NewClientArtist("Blur", ""), http.StatusBadRequest},
		"AddArtist: Non-admin":                                {"/admin/artist", "user", handlers.NewClientArtist("Blur", "link"), http.StatusUnauthorized},
		"AddArtist: Duplicate entry":                          {"/admin/artist", "admin", handlers.NewClientArtist(artist.Prefix+" "+artist.Name, linkArtist), http.StatusUnprocessableEntity},
		"AddArtist: Duplicate artist with different link":     {"/admin/artist", "admin", handlers.NewClientArtist(artist.Prefix+" "+artist.Name, "link"), http.StatusUnprocessableEntity},
		"AddArtist: Artist with same name but without prefix": {"/admin/artist", "admin", handlers.NewClientArtist(artist.Name, linkArtist), http.StatusUnprocessableEntity},
		"AddSong: Request without body":                       {"/admin/song", "admin", nil, http.StatusBadRequest},
		"AddSong: Existing artist with new song":              {"/admin/song", "admin", handlers.NewClientSong("No Good", artist.Name), http.StatusOK},
		"AddSong: New artist with new song":                   {"/admin/song", "admin", handlers.NewClientSong("Reason to Believe", "Sum 41"), http.StatusOK},
		"AddSong: Same song from other artist":                {"/admin/song", "admin", handlers.NewClientSong(song, "KDrew"), http.StatusOK},
		"AddSong: Collaboration of 2 new artists":             {"/admin/song", "admin", handlers.NewClientSong("Crazy", "Lost Frequencies", "Zonderling"), http.StatusOK},
		"AddSong: Collaboration of existing and new artist":   {"/admin/song", "admin", handlers.NewClientSong("Get Money", "Boogz Boogetz", artist.Name), http.StatusOK},
		"AddSong: Song without a name":                        {"/admin/song", "admin", handlers.NewClientSong("", "The Prodigy"), http.StatusBadRequest},
		"AddSong: Song without an artist":                     {"/admin/song", "admin", handlers.NewClientSong("House of the Rising Sun"), http.StatusBadRequest},
		"AddSong: Non-admin":                                  {"/admin/song", "user", handlers.NewClientSong("House of the Rising Sun", "The Animals"), http.StatusUnauthorized},
		"AddSong: Duplicate entry":                            {"/admin/song", "admin", handlers.NewClientSong(song, artist.Name), http.StatusUnprocessableEntity},
	}
	for name, test := range cases {
		db := newTestDB()
		if _, err := db.AddArtist(artist.Name, artist.Prefix, linkArtist); err != nil {
			t.Fatalf("%v: Failed to add artist for test TestAdminHandlers_response due to: %s\n", name, err)
			continue
		}
		if _, err := db.AddSong(song, []general.Artist{artist}); err != nil {
			t.Fatalf("%v: Failed to add song for test TestAdminHandlers_response due to: %s\n", name, err)
			continue
		}
		server, _ := testServerNoRequest(t, db)
		token, err := general.CreateToken(1, "test", test.roleClient)
		if err != nil {
			t.Errorf("Can't start TestAdminHandlers_response due to failure making token:%s\n", err)
			continue
		}
		response := general.TestRequest(t, server, http.MethodPost, test.path, token, test.body)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
	}
}

func TestAddArtistHandler_savingInDB(t *testing.T) {
	cases := map[string]struct {
		artist                       string
		expectedSavedInDB            bool
		expectedPrefix, expectedName string
	}{
		"Artist with a prefix":                                   {"The Rolling Stones", true, "The", "Rolling Stones"},
		"Artist with no prefix":                                  {"Sum 41", true, "", "Sum 41"},
		"Artist without prefix but the name starts with one (A)": {"Avenged Sevenfold", true, "", "Avenged Sevenfold"},
		"Empty artist name":                                      {"", false, "", ""},
	}
	token, err := general.CreateToken(1, "test", "admin")
	if err != nil {
		t.Fatalf("Can't start TestAddArtistHandler_savingInDB due to:%s\n", err)
	}
	for name, test := range cases {
		db := newTestDB()
		server, _ := testServerNoRequest(t, db)
		general.TestRequest(t, server, http.MethodPost, "/admin/artist", token, handlers.NewClientArtist(test.artist, "link"))
		result, ok := db.artistsDB[test.expectedName]
		if ok != test.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved in db: %v but got: %v\n", name, test.expectedSavedInDB, ok)
			continue
		}
		if result.name != test.expectedName {
			t.Errorf("%v: AddArtist %v expects name: %v but got: %v\n", name, test.artist, test.expectedName, result.name)
		}
		if result.prefix != test.expectedPrefix {
			t.Errorf("%v: AddArtist %v expects prefix: %v but got: %v\n", name, test.artist, test.expectedPrefix, result.prefix)
		}
	}
}

func TestAddSongHandler_savingInDB(t *testing.T) {
	cases := map[string]struct {
		artists           []string
		song              string
		artistToCheck     string
		expectedSavedInDB bool
	}{
		"Song with 1 artist":         {[]string{"Sum 41"}, "Reason to Believe", "Sum 41", true},
		"Song with multiple artists": {[]string{"Iggy Pop", "Sum 41"}, "Little Know It All", "Sum 41", true},
		"Song with no artist":        {[]string{""}, "No Good", "", false},
		"Song with no name":          {[]string{"Pendulum"}, "", "Pendulum", false},
	}
	token, err := general.CreateToken(1, "test", "admin")
	if err != nil {
		t.Fatalf("Can't start TestAddSongHandler_savingInDB due to:%s\n", err)
	}
	for name, test := range cases {
		db := newTestDB()
		server, _ := testServerNoRequest(t, db)
		general.TestRequest(t, server, http.MethodPost, "/admin/song", token, handlers.NewClientSong(test.song, test.artists...))
		song, ok := db.songsDB[test.artistToCheck][test.song]
		if ok != test.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved in db: %v but got: %v\n", name, test.expectedSavedInDB, ok)
			continue
		}
		if ok {
			if song.Name != test.song {
				t.Errorf("%v: AddSong expects name: %v but got: %v\n", name, test.song, song.Name)
			}
			if len(song.Artists) != len(test.artists) {
				t.Errorf("%v: AddSong expects %v artists but got: %v\n", name, len(test.artists), len(song.Artists))
			}
		}
	}
}

func TestAddArtist(t *testing.T) {
	artist := general.NewArtist(1, "Prodigy", "The")
	cases := map[string]struct {
		artist, prefix, link string
		expectedError        error
		expectedFoundInDB    bool
	}{
		"Artist with prefix":    {"Day to Remember", "A", "link", nil, true},
		"Artist without name":   {"", "A", "link", general.GetDBError("Missing name", general.InvalidInput), false},
		"Artist without prefix": {"Eminem", "", "link", nil, true},
		"Artist without link":   {"Metallica", "", "", nil, true},
		"Duplicate entry":       {"Prodigy", "The", "link", general.GetDBError("Duplicate entry", general.DuplicateEntry), true},
	}
	for name, test := range cases {
		db := newTestDB()
		if _, err := db.AddArtist(artist.Name, artist.Prefix, "link"); err != nil {
			t.Errorf("%v: Failed to set up test due to: %s\n", name, err)
			continue
		}
		handler, _ := testMusicHandlerNoRequest(t, db)
		_, err := handler.AddNewArtist(test.artist, test.prefix, test.link)
		if err == nil && test.expectedError != nil {
			t.Errorf("%v: Expects error with code %v but got no error\n", name, test.expectedError.(general.DBError).ErrorCode)
		}
		if err != nil && test.expectedError == nil {
			t.Errorf("%v: Expects no error but got error with code %v\n", name, err.(general.DBError).ErrorCode)
		}
		if err != nil && test.expectedError != nil && err.(general.DBError).ErrorCode != test.expectedError.(general.DBError).ErrorCode {
			t.Errorf("%v: Expects error with code %v but got error with code %v\n", name, test.expectedError.(general.DBError).ErrorCode, err.(general.DBError).ErrorCode)
		}
		if _, ok := db.artistsDB[test.artist]; ok != test.expectedFoundInDB {
			t.Errorf("%v: Expects to found artist in db is %v but got: %v\n", name, test.expectedFoundInDB, ok)
		}
	}
}

func TestAddArtist_sendMessage(t *testing.T) {
	artist := general.NewArtist(1, "Prodigy", "The")
	linkArtist := "linkToArtist"
	topic := "newArtist"
	cases := map[string]struct {
		artist, prefix, link string
		expectedFoundTopic   bool
	}{
		"Artist with prefix":    {"Day to Remember", "A", "link", true},
		"Artist without name":   {"", "A", "link", false},
		"Artist without prefix": {"Eminem", "", "link", true},
		"Artist without link":   {"Metallica", "", "", true},
		"Duplicate entry":       {"Prodigy", "The", "link", false},
	}
	for name, test := range cases {
		db := newTestDB()
		if _, err := db.AddArtist(artist.Name, artist.Prefix, linkArtist); err != nil {
			t.Errorf("%v: Failed to run test due to failing adding existing artist: %s\n", name, err)
			continue
		}
		handler, channel := testMusicHandlerNoRequest(t, db)
		handler.AddNewArtist(test.artist, test.prefix, test.link)
		go func() {
			time.Sleep(time.Millisecond)
			close(channel)
		}()
		foundTopic := false
		for message := range channel {
			if message.Topic != topic {
				t.Errorf("%v: Expects no other topic than %v but got topic: %v\n", name, topic, message.Topic)
				continue
			}
			foundTopic = true
			var result general.Artist
			if err := general.FromJSONBytes(&result, []byte(message.Message)); err != nil {
				t.Errorf("%v: Expects to send a message containing an artist but deserializing results in: %v\n", name, err)
				continue
			}
			if result.ID == 0 {
				t.Errorf("%v: Expects message with id but got id=0\n", name)
			}
			if result.Name != test.artist {
				t.Errorf("%v: Expects message with artist %v but got: %v\n", name, test.artist, result.Name)
			}
			if result.Prefix != test.prefix {
				t.Errorf("%v: Expects message with prefix %v but got: %v\n", name, test.prefix, result.Prefix)
			}
		}
		if foundTopic != test.expectedFoundTopic {
			t.Errorf("%v: Expects to found topic %v but got: %v\n", name, test.expectedFoundTopic, foundTopic)
		}
	}
}

func TestAddSong(t *testing.T) {
	artist := general.NewArtist(1, "Prodigy", "The")
	song := "No Good"
	cases := map[string]struct {
		artists           []string
		song              string
		artistToCheck     string
		expectedError     error
		expectedFoundInDB bool
	}{
		"New artist with new song":                                       {[]string{"Sum 41"}, "Fatlip", "Sum 41", nil, true},
		"Existing artist with new song":                                  {[]string{artist.Name}, "Warriors Dance", artist.Name, nil, true},
		"Other artist with same song name":                               {[]string{"Kaleo"}, song, "Kaleo", nil, true},
		"Collaboration of two new artists added to main artist":          {[]string{"Lost Frequencies", "Zonderling"}, "Crazy", "Lost Frequencies", nil, true},
		"Collaboration of two new artists added to collaborating artist": {[]string{"Lost Frequencies", "Zonderling"}, "Crazy", "Zonderling", nil, true},
		"Collaboration of existing and new artist":                       {[]string{artist.Name, "Boogz Boogetz"}, "Get Money", artist.Name, nil, true},
		"Song without artist":                                            {nil, "Still Waiting", "", general.GetDBError("Missing input", general.InvalidInput), false},
		"Song of existing artist without name of the song":               {[]string{artist.Name}, "", artist.Name, general.GetDBError("Missing input", general.InvalidInput), false},
		"Song of a new artist without the name of the song":              {[]string{"Sum 41"}, "", "Sum 41", general.GetDBError("Missing input", general.InvalidInput), false},
		"Song gets only added to collaborating artists":                  {[]string{"The Roling Stones"}, "Sympathy for the Devil", artist.Name, nil, false},
		"Duplicate song":                                                 {[]string{artist.Name}, song, artist.Name, general.GetDBError("Duplicate entry", general.DuplicateEntry), true},
	}
	for name, test := range cases {
		db := newTestDB()
		newArtist, err := db.AddArtist(artist.Name, artist.Prefix, "link")
		if err != nil {
			t.Errorf("%v: Failed to set up test with existing artist due to: %s\n", name, err)
			continue
		}
		if _, err = db.AddSong(song, []general.Artist{newArtist}); err != nil {
			t.Errorf("%v: Failed to set up test with existing song due to: %s\n", name, err)
			continue
		}
		handler, _ := testMusicHandlerNoRequest(t, db)
		_, err = handler.AddSong(test.song, test.artists...)
		if err == nil && test.expectedError != nil {
			t.Errorf("%v: Expects error with code %v but got no error\n", name, test.expectedError.(general.DBError).ErrorCode)
		}
		if err != nil && test.expectedError == nil {
			t.Errorf("%v: Expects no error but got error with code %v\n", name, err.(general.DBError).ErrorCode)
		}
		if err != nil && test.expectedError != nil && err.(general.DBError).ErrorCode != test.expectedError.(general.DBError).ErrorCode {
			t.Errorf("%v: Expects error with code %v but got error with code %v\n", name, test.expectedError.(general.DBError).ErrorCode, err.(general.DBError).ErrorCode)
		}
		if _, ok := db.songsDB[test.artistToCheck][test.song]; ok != test.expectedFoundInDB {
			t.Errorf("%v: Expects to found song in db is %v but got: %v\n", name, test.expectedFoundInDB, ok)
		}
	}
}

func TestAddSong_sendMessage(t *testing.T) {
	artist := general.NewArtist(1, "Queen", "")
	song := "Bohemian Rhapsody"
	topicSong := "newSong"
	cases := map[string]struct {
		artists                  []string
		song                     string
		topic                    string
		expectedFoundTopic       bool
		expectedFoundOtherTopics bool
	}{
		"Song from new artist":                                             {[]string{"Blur"}, "Song 2", "newSong", true, true},
		"Song from new artist also send newArtist message":                 {[]string{"Blur"}, "Song 2", "newArtist", true, true},
		"Song from existing artist":                                        {[]string{artist.Name}, "Doom and Gloom", "newSong", true, false},
		"Collaboration of existing and new artist":                         {[]string{artist.Name, "David Bowie"}, "Under Pressure", "newSong", true, true},
		"Collaboration of existing and new artist sends newArtist message": {[]string{artist.Name, "David Bowie"}, "Under Pressure", "newArtist", true, true},
		"Song without artist":                                              {nil, "Still Waiting", "newSong", false, false},
		"Song of existing artist without name of the song":                 {[]string{artist.Name}, "", "newSong", false, false},
		"Song of a new artist without the name of the song":                {[]string{"Sum 41"}, "", "newSong", false, true},
		"Duplicate song":                                                   {[]string{artist.Name}, song, "newSong", false, false},
	}
	for name, test := range cases {
		db := newTestDB()
		newArtist, err := db.AddArtist(artist.Name, artist.Prefix, "link")
		if err != nil {
			t.Errorf("%v: Failed to set up test with existing artist due to: %s\n", name, err)
			continue
		}
		if _, err = db.AddSong(song, []general.Artist{newArtist}); err != nil {
			t.Errorf("%v: Failed to set up test with existing song due to: %s\n", name, err)
			continue
		}
		handler, channel := testMusicHandlerNoRequest(t, db)
		handler.AddSong(test.song, test.artists...)
		go func() {
			time.Sleep(time.Millisecond)
			close(channel)
		}()
		foundTopic := false
		for message := range channel {
			if message.Topic != test.topic {
				if !test.expectedFoundOtherTopics {
					t.Errorf("%v: Expects no other topic than %v but got topic: %v\n", name, test.topic, message.Topic)
				}
				continue
			}
			foundTopic = true
			if message.Topic != topicSong {
				continue
			}
			var result general.Song
			if err := general.FromJSONBytes(&result, []byte(message.Message)); err != nil {
				t.Errorf("%v: Expects to send a message containing a song but deserializing results in: %v\n", name, err)
				continue
			}
			if result.ID == 0 {
				t.Errorf("%v: Expects message with id but got id=0\n", name)
			}
			if result.Name != test.song {
				t.Errorf("%v: Expects message with song %v but got: %v\n", name, test.song, result.Name)
			}
			if len(result.Artists) != len(test.artists) {
				t.Errorf("%v: Expects message with prefix %v but got: %v\n", name, len(test.artists), len(result.Artists))
			}
		}
		if foundTopic != test.expectedFoundTopic {
			t.Errorf("%v: Expects to found topic %v but got: %v\n", name, test.expectedFoundTopic, foundTopic)
		}
	}
}

package test

import (
	"discography/handlers"
	"general/convert"
	"general/testhelpers"
	"general/types"
	"net/http"
	"testing"
)

func TestUsersHandlers_response(t *testing.T) {
	// The discography is ordered by name of artist
	discography := map[string][]handlers.ClientSong{
		"30 Seconds to Mars": {handlers.NewClientSong("Kings and Queens", "30 Seconds to Mars")},
		"50 Cent":            {handlers.NewClientSong("Back Down", "50 Cent")},
		"David Bowie":        {handlers.NewClientSong("Space Oddity", "David Bowie")},
		"David Dallas":       {handlers.NewClientSong("Runnin'", "David Dallas")},
		"Disturbed":          {handlers.NewClientSong("The Sound of Silence", "Disturbed"), handlers.NewClientSong("Shout 2000", "Disturbed"), handlers.NewClientSong("Stricken", "Disturbed"), handlers.NewClientSong("Voices", "Disturbed")},
		"Doe Maar":           {handlers.NewClientSong("Doris Day", "Doe Maar")},
		"Evanescence":        {handlers.NewClientSong("Going Under", "Evanescence")},
		"Lost Frequencies":   {handlers.NewClientSong("Crazy", "Lost Frequencies", "Zonderling"), handlers.NewClientSong("Reality", "Lost Frequencies")},
		"The Offspring":      {handlers.NewClientSong("The Kids Aren't Alright", "The Offspring")},
		"Zonderling":         {},
	}
	cases := map[string]struct {
		path                  string
		expectedStatusCode    int
		expectedAmountResults int
		expectedHasNext       bool
	}{
		"ArtistStartingWith: Artist with startletter is found":                     {"/api/artists/L", http.StatusOK, 1, false},
		"ArtistStartingWith: Artist with startnumber is found":                     {"/api/artists/3", http.StatusOK, 1, false},
		"ArtistStartingWith: All artist with startnumber is found":                 {"/api/artists/0-9", http.StatusOK, 2, false},
		"ArtistStartingWith: Only first letter counts":                             {"/api/artist/B", http.StatusNotFound, 0, false},
		"ArtistStartingWith: Only first number counts":                             {"/api/artist/0", http.StatusNotFound, 0, false},
		"ArtistStartingWith: Artist with prefix and correct startletter":           {"/api/artists/O", http.StatusOK, 1, false},
		"ArtistStartingWith: Prefix doesn't count as startletter":                  {"/api/artists/T", http.StatusNotFound, 0, false},
		"ArtistStartingWith: No artist is found":                                   {"/api/artists/X", http.StatusNotFound, 0, false},
		"ArtistStartingWith: Multiple artists are found":                           {"/api/artists/D", http.StatusOK, 4, false},
		"ArtistStartingWith: Offset is bigger than amount of results":              {"/api/artists/D?offset=100", http.StatusNotFound, 0, false},
		"ArtistStartingWith: Skip amount of results given by offset":               {"/api/artists/D?offset=1", http.StatusOK, 3, false},
		"ArtistStartingWith: Amount of results capped by max":                      {"/api/artists/D?max=3", http.StatusOK, 3, true},
		"ArtistStartingWith: Max is bigger than amount of results":                 {"/api/artists/D?max=100", http.StatusOK, 4, false},
		"ArtistStartingWith: Sum offset and max is smaller than amount of results": {"/api/artists/D?offset=1&max=2", http.StatusOK, 2, true},
		"ArtistStartingWith: Sum offset and max is equal to amount of results":     {"/api/artists/D?offset=2&max=2", http.StatusOK, 2, false},
		"ArtistStartingWith: Sum offset and max is bigger than amount of results":  {"/api/artists/D?offset=2&max=3", http.StatusOK, 2, false},
		"SongsFromArtist: Artist has some songs":                                   {"/api/artist/Evanescence", http.StatusOK, 1, false},
		"SongsFromArtist: Songs from artist with space are found":                  {"/api/artist/David%20Bowie", http.StatusOK, 1, false},
		"SongsFromArtist: Songs from artist with prefix are found":                 {"/api/artist/Offspring", http.StatusOK, 1, false},
		"SongsFromArtist: No song is found":                                        {"/api/artist/Dio", http.StatusNotFound, 0, false},
		"SongsFromArtist: Artist has multiple songs":                               {"/api/artist/Disturbed", http.StatusOK, 4, false},
		"SongsFromArtist: Collaborations as main artist are included":              {"/api/artist/Lost%20Frequencies", http.StatusOK, 2, false},
		"SongsFromArtist: Collaborations as non-leading artist are included":       {"/api/artist/Zonderling", http.StatusOK, 1, false},
		"SongsFromArtist: Offset is bigger than amount of results":                 {"/api/artist/Disturbed?offset=100", http.StatusNotFound, 0, false},
		"SongsFromArtist: Skip amount of results given by offset":                  {"/api/artist/Disturbed?offset=1", http.StatusOK, 3, false},
		"SongsFromArtist: Amount of results capped by max":                         {"/api/artist/Disturbed?max=3", http.StatusOK, 3, true},
		"SongsFromArtist: Max is bigger than amount of results":                    {"/api/artist/Disturbed?max=100", http.StatusOK, 4, false},
		"SongsFromArtist: Sum offset and max is smaller than amount of results":    {"/api/artist/Disturbed?offset=1&max=2", http.StatusOK, 2, true},
		"SongsFromArtist: Sum offset and max is equal to amount of results":        {"/api/artist/Disturbed?offset=2&max=2", http.StatusOK, 2, false},
		"SongsFromArtist: Sum offset and max is bigger than amount of results":     {"/api/artist/Disturbed?offset=2&max=3", http.StatusOK, 2, false},
	}
	for name, test := range cases {
		db := newTestDB()
		if err := testAddDiscographyToDB(t, db, discography); err != nil {
			t.Fatalf("Can't add song for test TestAdminHandlers_response due to: %s\n", err)
			continue
		}
		testServer, _ := testServerNoRequest(t, db)
		response := testhelpers.TestRequest(t, testServer, http.MethodGet, test.path, "", nil)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
		if response.Code != http.StatusOK {
			continue
		}
		var results types.Music
		if err := convert.ReadFromJSON(&results, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", name, err)
			continue
		}
		if len(results.Data) != test.expectedAmountResults {
			t.Errorf("%v: Expect to find %v results, but got: %v\n", name, test.expectedAmountResults, len(results.Data))
		}
		if results.HasNext != test.expectedHasNext {
			t.Errorf("%v: Expects HasNext is %v but got: %v\n", name, test.expectedHasNext, results.HasNext)
		}
	}
}

func TestArtistStartingWith_orderResults(t *testing.T) {
	// The artists are ordered by name but not by id.
	artists := []types.Artist{types.NewArtist(12, "Beatles", "The"), types.NewArtist(2, "Blur", ""), types.NewArtist(21, "Bob Dylan", ""), types.NewArtist(3, "Bon Jovi", "")}
	cases := map[string]struct {
		path              string
		indexExpectedName int
		expectedName      string
	}{
		"Results are ordered":                          {"/api/artists/B", 1, artists[1].Name},
		"Correct first result":                         {"/api/artists/B", 0, artists[0].Name},
		"Correct last result":                          {"/api/artists/B", len(artists) - 1, artists[len(artists)-1].Name},
		"Skip the right artists when offset is given":  {"/api/artists/B?offset=2", 0, artists[2].Name},
		"Stops at the right artists when max is given": {"/api/artists/B?max=3", 2, artists[2].Name},
	}
	for name, test := range cases {
		db := newTestDB()
		for _, artist := range artists {
			if _, err := db.AddArtist(artist.Name, artist.Prefix, "link"); err != nil {
				t.Fatalf("Can't start test TestArtistStartingWith_orderResults due to failure of adding artist %v:%s\n", artist.Name, err)
			}
		}
		testServer, _ := testServerNoRequest(t, db)
		response := testhelpers.TestRequest(t, testServer, http.MethodGet, test.path, "", nil)
		if response.Code != http.StatusOK {
			t.Errorf("%v: Expects statuscode %v but got: %v\n", name, http.StatusOK, response.Code)
			continue
		}
		var result types.MultipleArtists
		if err := convert.ReadFromJSON(&result, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", name, err)
			continue
		}
		if len(result.Data) <= test.indexExpectedName {
			t.Errorf("%v: Expects at least %v results, but got: %v\n", name, test.indexExpectedName+1, len(result.Data))
			continue
		}
		if artist := result.Data[test.indexExpectedName].Name; artist != test.expectedName {
			t.Errorf("%v: Searching for artists with given first letter expects %v at index %v but got %v\n", name, test.expectedName, test.indexExpectedName, artist)
		}
	}
}

func TestSongsFromArtist_orderResults(t *testing.T) {
	// The songs are ordered by name.
	songsBonJovi := []string{"Always", "Bed of Roses", "It's My Life", "Living on a Prayer"}
	cases := map[string]struct {
		path              string
		indexExpectedName int
		expectedName      string
	}{
		"Results are ordered":                          {"/api/artist/Bon%20Jovi", 1, songsBonJovi[1]},
		"Correct first result":                         {"/api/artist/Bon%20Jovi", 0, songsBonJovi[0]},
		"Correct last result":                          {"/api/artist/Bon%20Jovi", len(songsBonJovi) - 1, songsBonJovi[len(songsBonJovi)-1]},
		"Skip the right artists when offset is given":  {"/api/artist/Bon%20Jovi?offset=2", 0, songsBonJovi[2]},
		"Stops at the right artists when max is given": {"/api/artist/Bon%20Jovi?max=3", 2, songsBonJovi[2]},
	}
	for name, test := range cases {
		db := newTestDB()
		artist, err := db.AddArtist("Bon Jovi", "", "link")
		if err != nil {
			t.Fatalf("Can't start test TestSongsFromArtist_orderResults due to failure of adding artist Bon Jovi:%s\n", err)
		}
		for _, song := range songsBonJovi {
			if _, err := db.AddSong(song, []types.Artist{artist}); err != nil {
				t.Fatalf("Can't start test TestSongsFromArtist_correctOrderResults due to failure of adding song %v:%s\n", song, err)
			}
		}
		testServer, _ := testServerNoRequest(t, db)
		response := testhelpers.TestRequest(t, testServer, http.MethodGet, test.path, "", nil)
		if response.Code != http.StatusOK {
			t.Errorf("%v: Expects statuscode %v but got: %v\n", name, http.StatusOK, response.Code)
			continue
		}
		var result types.MultipleSongs
		if err = convert.ReadFromJSON(&result, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", name, err)
			continue
		}
		if len(result.Data) <= test.indexExpectedName {
			t.Errorf("%v: Expects at least %v results, but got: %v\n", name, test.indexExpectedName+1, len(result.Data))
			continue
		}
		if artist := result.Data[test.indexExpectedName].Name; artist != test.expectedName {
			t.Errorf("%v: Searching for songs of artist expects %v at index %v but got %v\n", name, test.expectedName, test.indexExpectedName, artist)
		}
	}
}

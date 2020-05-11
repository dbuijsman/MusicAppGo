package test

import (
	"general"
	"net/http"
	"testing"
)

func TestGetHandlers_response(t *testing.T) {
	user := general.NewCredentials(1, "Test", "user")
	userWithNoPrefs := general.NewCredentials(2, "NoPrefs", "user")
	missingUser := general.NewCredentials(3, "Missing", "user")
	artists := map[string]general.Artist{
		"Sum 41":   general.NewArtist(1, "Sum 41", ""),
		"Slipknot": general.NewArtist(2, "Slipknot", ""),
		"ZZ Top":   general.NewArtist(3, "ZZ Top", ""),
	}
	likedSongs := []general.Song{general.NewSong(11, []general.Artist{artists["Sum 41"]}, "In Too Deep"), general.NewSong(12, []general.Artist{artists["Sum 41"]}, "Walking Disaster"), general.NewSong(13, []general.Artist{artists["Slipknot"]}, "Duality"), general.NewSong(14, []general.Artist{artists["Slipknot"]}, "Snuff")}
	dislikedSongs := []general.Song{general.NewSong(21, []general.Artist{artists["ZZ Top"]}, "Viva Las Vegas"), general.NewSong(22, []general.Artist{artists["ZZ Top"]}, "I Gotsta Get Paid"), general.NewSong(23, []general.Artist{artists["ZZ Top"]}, "La Grange"), general.NewSong(24, []general.Artist{artists["Slipknot"]}, "Nero Forte")}
	otherSongs := []general.Song{general.NewSong(31, []general.Artist{artists["Sum 41"]}, "Reason to Believe")}
	cases := map[string]struct {
		path                  string
		credentials           *general.Credentials
		expectedStatusCode    int
		expectedAmountResults int
		expectedHasNext       bool
	}{
		"GetLikes: User has likes":                                          {"/api/like", &user, http.StatusOK, 4, false},
		"GetLikes: User has no likes":                                       {"/api/like", &userWithNoPrefs, http.StatusNotFound, 0, false},
		"GetLikes: User is missing":                                         {"/api/like", &missingUser, http.StatusNotFound, 0, false},
		"GetLikes: No token is given":                                       {"/api/like", nil, http.StatusUnauthorized, 0, false},
		"GetLikes: Offset is bigger than amount of results":                 {"/api/like?offset=100", &user, http.StatusNotFound, 3, false},
		"GetLikes: Skip results given by offset":                            {"/api/like?offset=1", &user, http.StatusOK, 3, false},
		"GetLikes: Amount is capped by max":                                 {"/api/like?max=3", &user, http.StatusOK, 3, true},
		"GetLikes: Max is bigger than amount of results":                    {"/api/like?max=100", &user, http.StatusOK, 4, false},
		"GetLikes: Sum offset and max is smaller than amount of results":    {"/api/like?offset=1&max=2", &user, http.StatusOK, 2, true},
		"GetLikes: Sum offset and max is equal to amount of results":        {"/api/like?offset=2&max=2", &user, http.StatusOK, 2, false},
		"GetLikes: Sum offset and max is bigger than amount of results":     {"/api/like?offset=2&max=3", &user, http.StatusOK, 2, false},
		"GetDislikes: User has dislikes":                                    {"/api/dislike", &user, http.StatusOK, 4, false},
		"GetDislikes: User has no dislikes":                                 {"/api/dislike", &userWithNoPrefs, http.StatusNotFound, 0, false},
		"GetDislikes: User is missing":                                      {"/api/dislike", &missingUser, http.StatusNotFound, 0, false},
		"GetDislikes: No token is given":                                    {"/api/dislike", nil, http.StatusUnauthorized, 0, false},
		"GetDislikes: Offset is bigger than amount of results":              {"/api/dislike?offset=100", &user, http.StatusNotFound, 3, false},
		"GetDislikes: Skip results given by offset":                         {"/api/dislike?offset=1", &user, http.StatusOK, 3, false},
		"GetDislikes: Amount is capped by max":                              {"/api/dislike?max=3", &user, http.StatusOK, 3, true},
		"GetDislikes: Max is bigger than amount of results":                 {"/api/dislike?max=100", &user, http.StatusOK, 4, false},
		"GetDislikes: Sum offset and max is smaller than amount of results": {"/api/dislike?offset=1&max=2", &user, http.StatusOK, 2, true},
		"GetDislikes: Sum offset and max is equal to amount of results":     {"/api/dislike?offset=2&max=2", &user, http.StatusOK, 2, false},
		"GetDislikes: Sum offset and max is bigger than amount of results":  {"/api/dislike?offset=2&max=3", &user, http.StatusOK, 2, false},
	}
	for name, test := range cases {
		db := newTestDB()
		if err := db.AddUser(user); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		if err := db.AddUser(userWithNoPrefs); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		db.addPreferencesToTestDB(t, user.ID, likedSongs, db.AddLike)
		db.addPreferencesToTestDB(t, user.ID, dislikedSongs, db.AddDislike)
		db.addSongsToTestDB(t, otherSongs)
		server := testServer(db, addDBToArray(make([]general.Song, 0), db))
		token := ""
		if test.credentials != nil {
			var err error
			token, err = general.CreateToken(test.credentials.ID, test.credentials.Username, test.credentials.Role)
			if err != nil {
				t.Errorf("%v: Failed to send token with request: %s\n", name, err)
				continue
			}
		}
		response := general.TestRequest(t, server, http.MethodGet, test.path, token, nil)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
		if response.Code != http.StatusOK {
			continue
		}
		var results general.MultipleSongs
		if err := general.ReadFromJSON(&results, response.Body); err != nil {
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

func TestGetHandlers_orderResults(t *testing.T) {
	user := general.NewCredentials(1, "Test", "user")
	artists := map[string]general.Artist{
		"Iggy Pop":         general.NewArtist(1, "Iggy Pop", ""),
		"Lost Frequencies": general.NewArtist(2, "Lost Frequencies", ""),
		"Sum 41":           general.NewArtist(3, "Sum 41", ""),
		"Slipknot":         general.NewArtist(4, "Slipknot", ""),
		"Zonderling":       general.NewArtist(5, "Zonderling", ""),
		"ZZ Top":           general.NewArtist(6, "ZZ Top", ""),
	}
	// likedSongs and dislikedSongs are ordered by artist and then by name of the song. Artists of collaborations are sorted in reverse order
	likedSongs := []general.Song{general.NewSong(15, []general.Artist{artists["Iggy Pop"]}, "I'm Bored"), general.NewSong(16, []general.Artist{artists["Sum 41"], artists["Iggy Pop"]}, "Little Know It All"), general.NewSong(13, []general.Artist{artists["Slipknot"]}, "Duality"), general.NewSong(14, []general.Artist{artists["Slipknot"]}, "Snuff"), general.NewSong(11, []general.Artist{artists["Sum 41"]}, "In Too Deep"), general.NewSong(12, []general.Artist{artists["Sum 41"]}, "Walking Disaster")}
	dislikedSongs := []general.Song{general.NewSong(26, []general.Artist{artists["Lost Frequencies"]}, "Are You With Me"), general.NewSong(25, []general.Artist{artists["Zonderling"], artists["Lost Frequencies"]}, "Crazy"), general.NewSong(23, []general.Artist{artists["Slipknot"]}, "All Out Life"), general.NewSong(24, []general.Artist{artists["Slipknot"]}, "Nero Forte"), general.NewSong(22, []general.Artist{artists["ZZ Top"]}, "I Gotsta Get Paid"), general.NewSong(21, []general.Artist{artists["ZZ Top"]}, "Viva Las Vegas")}
	otherSongs := []general.Song{general.NewSong(31, []general.Artist{artists["Sum 41"]}, "Reason to Believe")}
	cases := map[string]struct {
		path                                           string
		indexResult                                    int
		expectedFirstArtist, expectedName, expectedTag string
	}{
		"GetLikes: Results are ordered":                                       {"/api/like", 4, "Sum 41", "In Too Deep", "like"},
		"GetLikes: Correct first result":                                      {"/api/like", 0, "Iggy Pop", "I'm Bored", "like"},
		"GetLikes: Correct last result":                                       {"/api/like", 5, "Sum 41", "Walking Disaster", "like"},
		"GetLikes: Orders first by artist, then song":                         {"/api/like", 2, "Slipknot", "Duality", "like"},
		"GetLikes: Collaborations takes first ordered artist for ordering":    {"/api/like", 1, "Iggy Pop", "Little Know It All", "like"},
		"GetLikes: Skip the right songs when offset is given":                 {"/api/like?offset=2", 0, "Slipknot", "Duality", "like"},
		"GetLikes: Stops ar the right artist when max is given":               {"/api/like?max=4", 3, "Slipknot", "Snuff", "like"},
		"GetDislikes: Results are ordered":                                    {"/api/dislike", 4, "ZZ Top", "I Gotsta Get Paid", "dislike"},
		"GetDislikes: Correct first result":                                   {"/api/dislike", 0, "Lost Frequencies", "Are You With Me", "dislike"},
		"GetDislikes: Correct last result":                                    {"/api/dislike", 5, "ZZ Top", "Viva Las Vegas", "dislike"},
		"GetDislikes: Orders first by artist, then song":                      {"/api/dislike", 2, "Slipknot", "All Out Life", "dislike"},
		"GetDislikes: Collaborations takes first ordered artist for ordering": {"/api/dislike", 1, "Lost Frequencies", "Crazy", "dislike"},
		"GetDislikes: Skip the right songs when offset is given":              {"/api/dislike?offset=2", 0, "Slipknot", "All Out Life", "dislike"},
		"GetDislikes: Stops ar the right artist when max is given":            {"/api/dislike?max=4", 3, "Slipknot", "Nero Forte", "dislike"},
	}
	token, err := general.CreateToken(1, "test", "admin")
	if err != nil {
		t.Fatalf("Can't start TestGetHandlers_orderResults due to failure creating token:%s\n", err)
	}
	for name, test := range cases {
		db := newTestDB()
		if err := db.AddUser(user); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		db.addPreferencesToTestDB(t, user.ID, likedSongs, db.AddLike)
		db.addPreferencesToTestDB(t, user.ID, dislikedSongs, db.AddDislike)
		db.addSongsToTestDB(t, otherSongs)
		server := testServer(db, addDBToArray(make([]general.Song, 0), db))
		response := general.TestRequest(t, server, http.MethodGet, test.path, token, nil)
		if response.Code != http.StatusOK {
			t.Errorf("%v: Expects statuscode %v but got: %v\n", name, http.StatusOK, response.Code)
			continue
		}
		var results general.MultipleSongs
		if err := general.ReadFromJSON(&results, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", name, err)
			continue
		}
		if len(results.Data) <= test.indexResult {
			t.Errorf("%v: Expects at least %v results but got: %v\n", name, test.indexResult+1, len(results.Data))
			continue
		}
		song := results.Data[test.indexResult]
		if artist := getFirstOrderedArtist(song); artist.Name != test.expectedFirstArtist {
			t.Errorf("%v: Expects to find song with artist %v at index %v but got: %v\n", name, test.expectedFirstArtist, test.indexResult, artist)
		}
		if song.Name != test.expectedName {
			t.Errorf("%v: Expects to find song %v at index %v but got: %v\n", name, test.expectedName, test.indexResult, song.Name)
		}
		if song.Preference != test.expectedTag {
			t.Errorf("%v: Expects to find song with tag %v at index %v but got: %v\n", name, test.expectedTag, test.indexResult, song.Preference)
		}
	}

}

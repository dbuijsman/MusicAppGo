package test

import (
	"general"
	"net/http"
	"testing"
	"time"
)

func TestChangeHandlers_statusCode(t *testing.T) {
	user := general.NewCredentials(1, "Test", "user")
	missingUser := general.NewCredentials(2, "Missing", "user")
	artists := map[string]general.Artist{
		"Sum 41":   general.NewArtist(1, "Sum 41", ""),
		"Slipknot": general.NewArtist(2, "Slipknot", ""),
		"ZZ Top":   general.NewArtist(3, "ZZ Top", ""),
	}
	likedSongs := []general.Song{general.NewSong(11, []general.Artist{artists["Sum 41"]}, "In Too Deep"), general.NewSong(12, []general.Artist{artists["Slipknot"]}, "Duality")}
	dislikedSongs := []general.Song{general.NewSong(21, []general.Artist{artists["ZZ Top"]}, "Viva Las Vegas")}
	otherSongs := []general.Song{general.NewSong(31, []general.Artist{artists["Sum 41"]}, "Reason to Believe")}
	missingSongs := []general.Song{general.NewSong(41, []general.Artist{artists["Sum 41"]}, "Happiness Machine")}
	cases := map[string]struct {
		method, path       string
		credentials        *general.Credentials
		body               interface{}
		expectedStatusCode int
	}{
		"AddLike: Add a new like":                               {http.MethodPost, "/api/like", &user, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"AddLike: Change dislike to like":                       {http.MethodPost, "/api/like", &user, general.NewPreference(dislikedSongs[0].ID, "artist"), http.StatusOK},
		"AddLike: Like a liked song":                            {http.MethodPost, "/api/like", &user, general.NewPreference(likedSongs[0].ID, "artist"), http.StatusOK},
		"AddLike: Add a new like for a missing user":            {http.MethodPost, "/api/like", &missingUser, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"AddLike: Add a new like for a missing song":            {http.MethodPost, "/api/like", &user, general.NewPreference(missingSongs[0].ID, "artist"), http.StatusOK},
		"AddLike: Add a new like for non-existing song":         {http.MethodPost, "/api/like", &user, general.NewPreference(404, "artist"), http.StatusNotFound},
		"AddLike: Sending no token":                             {http.MethodPost, "/api/like", nil, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusUnauthorized},
		"AddLike: Sending no songID":                            {http.MethodPost, "/api/like", &user, general.NewPreference(0, "artist"), http.StatusBadRequest},
		"AddLike: Sending no body":                              {http.MethodPost, "/api/like", &user, nil, http.StatusBadRequest},
		"AddDislike: Add a new dislike":                         {http.MethodPost, "/api/dislike", &user, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"AddDislike: Change like to dislike":                    {http.MethodPost, "/api/dislike", &user, general.NewPreference(likedSongs[0].ID, "artist"), http.StatusOK},
		"AddDislike: Dislike a disliked song":                   {http.MethodPost, "/api/dislike", &user, general.NewPreference(dislikedSongs[0].ID, "artist"), http.StatusOK},
		"AddDislike: Add a new dislike for a missing user":      {http.MethodPost, "/api/dislike", &missingUser, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"AddDislike: Add a new dislike for a missing song":      {http.MethodPost, "/api/dislike", &user, general.NewPreference(missingSongs[0].ID, "artist"), http.StatusOK},
		"AddDislike: Add a new dislike for non-existing song":   {http.MethodPost, "/api/dislike", &user, general.NewPreference(404, "artist"), http.StatusNotFound},
		"AddDislike: Sending no token":                          {http.MethodPost, "/api/dislike", nil, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusUnauthorized},
		"AddDislike: Sending no songID":                         {http.MethodPost, "/api/dislike", &user, general.NewPreference(0, "artist"), http.StatusBadRequest},
		"AddDislike: Sending no body":                           {http.MethodPost, "/api/dislike", &user, nil, http.StatusBadRequest},
		"RemoveLike: Delete a like":                             {http.MethodDelete, "/api/like", &user, general.NewPreference(likedSongs[0].ID, "artist"), http.StatusOK},
		"RemoveLike: Delete a like that is disliked":            {http.MethodDelete, "/api/like", &user, general.NewPreference(dislikedSongs[0].ID, "artist"), http.StatusOK},
		"RemoveLike: Delete a like that is no preference":       {http.MethodDelete, "/api/like", &user, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"RemoveLike: Delete a like that is missing":             {http.MethodDelete, "/api/like", &user, general.NewPreference(missingSongs[0].ID, "artist"), http.StatusOK},
		"RemoveLike: Remove a like for a missing user":          {http.MethodDelete, "/api/like", &missingUser, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"RemoveLike: Remove a like for non-existing song":       {http.MethodDelete, "/api/like", &user, general.NewPreference(404, "artist"), http.StatusOK},
		"RemoveLike: Sending no token":                          {http.MethodDelete, "/api/like", nil, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusUnauthorized},
		"RemoveLike: Sending no songID":                         {http.MethodDelete, "/api/like", &user, general.NewPreference(0, "artist"), http.StatusBadRequest},
		"RemoveLike: Sending no body":                           {http.MethodDelete, "/api/like", &user, nil, http.StatusBadRequest},
		"RemoveDislike: Delete a dislike":                       {http.MethodDelete, "/api/dislike", &user, general.NewPreference(dislikedSongs[0].ID, "artist"), http.StatusOK},
		"RemoveDislike: Delete a dislike that is liked":         {http.MethodDelete, "/api/dislike", &user, general.NewPreference(likedSongs[0].ID, "artist"), http.StatusOK},
		"RemoveDislike: Delete a dislike that is no preference": {http.MethodDelete, "/api/dislike", &user, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"RemoveDislike: Delete a dislike that is missing":       {http.MethodDelete, "/api/dislike", &user, general.NewPreference(missingSongs[0].ID, "artist"), http.StatusOK},
		"RemoveDislike: Remove a dislike for a missing user":    {http.MethodDelete, "/api/dislike", &missingUser, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusOK},
		"RemoveDislike: Remove a dislike for non-existing song": {http.MethodDelete, "/api/dislike", &user, general.NewPreference(404, "artist"), http.StatusOK},
		"RemoveDislike: Sending no token":                       {http.MethodDelete, "/api/dislike", nil, general.NewPreference(otherSongs[0].ID, "artist"), http.StatusUnauthorized},
		"RemoveDislike: Sending no songID":                      {http.MethodDelete, "/api/dislike", &user, general.NewPreference(0, "artist"), http.StatusBadRequest},
		"RemoveDislike: Sending no body":                        {http.MethodDelete, "/api/dislike", &user, nil, http.StatusBadRequest},
	}
	for name, test := range cases {
		db := newTestDB()
		if err := db.AddUser(user); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		db.addPreferencesToTestDB(t, user.ID, likedSongs, db.AddLike)
		db.addPreferencesToTestDB(t, user.ID, dislikedSongs, db.AddDislike)
		db.addSongsToTestDB(t, otherSongs)
		server := testServer(db, addDBToArray(missingSongs, db))
		token := ""
		if test.credentials != nil {
			var err error
			token, err = general.CreateToken(test.credentials.ID, test.credentials.Username, test.credentials.Role)
			if err != nil {
				t.Errorf("%v: Failed to send token with request: %s\n", name, err)
				continue
			}
		}
		response := general.TestRequest(t, server, test.method, test.path, token, test.body)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
	}
}

func TestChangeHandlers_savingInDB(t *testing.T) {
	user := general.NewCredentials(1, "Test", "user")
	missingUser := general.NewCredentials(2, "Missing", "user")
	artists := map[string]general.Artist{
		"Sum 41":   general.NewArtist(1, "Sum 41", ""),
		"Slipknot": general.NewArtist(2, "Slipknot", ""),
		"ZZ Top":   general.NewArtist(3, "ZZ Top", ""),
	}
	likedSongs := []general.Song{general.NewSong(11, []general.Artist{artists["Sum 41"]}, "In Too Deep"), general.NewSong(12, []general.Artist{artists["Slipknot"]}, "Duality")}
	dislikedSongs := []general.Song{general.NewSong(21, []general.Artist{artists["ZZ Top"]}, "Viva Las Vegas")}
	otherSongs := []general.Song{general.NewSong(31, []general.Artist{artists["Sum 41"]}, "Reason to Believe")}
	missingSongs := []general.Song{general.NewSong(41, []general.Artist{artists["Sum 41"]}, "Happiness Machine")}
	cases := map[string]struct {
		method, path                  string
		credentials                   *general.Credentials
		songID                        int
		expectedLike, expectedDislike bool
	}{
		"AddLike: Add a new like":                               {http.MethodPost, "/api/like", &user, otherSongs[0].ID, true, false},
		"AddLike: Change dislike to like":                       {http.MethodPost, "/api/like", &user, dislikedSongs[0].ID, true, false},
		"AddLike: Add a new like for a missing user":            {http.MethodPost, "/api/like", &missingUser, otherSongs[0].ID, true, false},
		"AddLike: Add a new like for a missing song":            {http.MethodPost, "/api/like", &user, missingSongs[0].ID, true, false},
		"AddDislike: Add a new dislike":                         {http.MethodPost, "/api/dislike", &user, otherSongs[0].ID, false, true},
		"AddDislike: Change like to dislike":                    {http.MethodPost, "/api/dislike", &user, likedSongs[0].ID, false, true},
		"AddDislike: Add a new dislike for a missing user":      {http.MethodPost, "/api/dislike", &missingUser, otherSongs[0].ID, false, true},
		"AddDislike: Add a new dislike for a missing song":      {http.MethodPost, "/api/dislike", &user, missingSongs[0].ID, false, true},
		"RemoveLike: Delete a like":                             {http.MethodDelete, "/api/like", &user, likedSongs[0].ID, false, false},
		"RemoveLike: Delete a like that is disliked":            {http.MethodDelete, "/api/like", &user, dislikedSongs[0].ID, false, true},
		"RemoveLike: Delete a like that is no preference":       {http.MethodDelete, "/api/like", &user, otherSongs[0].ID, false, false},
		"RemoveDislike: Delete a dislike":                       {http.MethodDelete, "/api/dislike", &user, dislikedSongs[0].ID, false, false},
		"RemoveDislike: Delete a dislike that is liked":         {http.MethodDelete, "/api/dislike", &user, likedSongs[0].ID, true, false},
		"RemoveDislike: Delete a dislike that is no preference": {http.MethodDelete, "/api/dislike", &user, otherSongs[0].ID, false, false},
	}
	for name, test := range cases {
		db := newTestDB()
		if err := db.AddUser(user); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		db.addPreferencesToTestDB(t, user.ID, likedSongs, db.AddLike)
		db.addPreferencesToTestDB(t, user.ID, dislikedSongs, db.AddDislike)
		db.addSongsToTestDB(t, otherSongs)
		server := testServer(db, addDBToArray(missingSongs, db))
		token, err := general.CreateToken(test.credentials.ID, test.credentials.Username, test.credentials.Role)
		if err != nil {
			t.Errorf("%v: Failed to send token with request: %s\n", name, err)
			continue
		}
		general.TestRequest(t, server, test.method, test.path, token, general.NewPreference(test.songID, "artist"))
		time.Sleep(time.Millisecond)
		if _, ok := db.likes[test.credentials.ID][test.songID]; ok != test.expectedLike {
			t.Errorf("%v: Expects to be a like is %v but got: %v\n", name, test.expectedLike, ok)
		}
		if _, ok := db.dislikes[test.credentials.ID][test.songID]; ok != test.expectedDislike {
			t.Errorf("%v: Expects to be a dislike is %v but got: %v\n", name, test.expectedDislike, ok)
		}
	}
}

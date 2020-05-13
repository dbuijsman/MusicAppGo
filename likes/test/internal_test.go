package test

import (
	"general/convert"
	"general/server"
	"general/testhelpers"
	"general/types"
	"net/http"
	"testing"
)

func TestInternal_response(t *testing.T) {
	user := types.NewCredentials(1, "Test", "user")
	var internal, userToken, adminToken string
	var err error
	if internal, err = server.CreateTokenInternalRequests("testServer"); err != nil {
		t.Fatalf("Failed to create internal token: %s\n", err)
	}
	if userToken, err = server.CreateToken(user.ID, user.Username, user.Role); err != nil {
		t.Fatalf("Failed to create user token: %s\n", err)
	}
	if adminToken, err = server.CreateToken(2, "testAdmin", "admin"); err != nil {
		t.Logf("%v, %v\n", userToken, adminToken)
		t.Fatalf("Failed to create admin token: %s\n", err)
	}
	artists := map[string]types.Artist{
		"Sum 41":   types.NewArtist(1, "Sum 41", ""),
		"Slipknot": types.NewArtist(2, "Slipknot", ""),
		"ZZ Top":   types.NewArtist(3, "ZZ Top", ""),
	}
	likedSongs := []types.Song{types.NewSong(11, []types.Artist{artists["Sum 41"]}, "In Too Deep"), types.NewSong(12, []types.Artist{artists["Sum 41"]}, "Walking Disaster"), types.NewSong(13, []types.Artist{artists["Slipknot"]}, "Duality"), types.NewSong(14, []types.Artist{artists["Slipknot"]}, "Snuff")}
	dislikedSongs := []types.Song{types.NewSong(21, []types.Artist{artists["ZZ Top"]}, "Viva Las Vegas"), types.NewSong(22, []types.Artist{artists["ZZ Top"]}, "I Gotsta Get Paid"), types.NewSong(23, []types.Artist{artists["ZZ Top"]}, "La Grange"), types.NewSong(24, []types.Artist{artists["Slipknot"]}, "Nero Forte")}
	otherSongs := []types.Song{types.NewSong(31, []types.Artist{artists["Sum 41"]}, "Reason to Believe")}
	cases := map[string]struct {
		path, token        string
		idExpectedTag      int
		expectedStatusCode int
		expectedTag        string
	}{
		"GetPreferenceOfArtist: Valid token and existing song":         {"/intern/preference/1/Sum%2041", internal, -1, http.StatusOK, ""},
		"GetPreferenceOfArtist: User token is not authorized":          {"/intern/preference/1/Sum%2041", userToken, 0, http.StatusUnauthorized, ""},
		"GetPreferenceOfArtist: Admin token is not authorized":         {"/intern/preference/1/Sum%2041", adminToken, 0, http.StatusUnauthorized, ""},
		"GetPreferenceOfArtist: No token is send":                      {"/intern/preference/1/Sum%2041", "", 0, http.StatusUnauthorized, ""},
		"GetPreferenceOfArtist: Includes likes from artist":            {"/intern/preference/1/Slipknot", internal, 13, http.StatusOK, "like"},
		"GetPreferenceOfArtist: Includes dislikes from artist":         {"/intern/preference/1/Slipknot", internal, 24, http.StatusOK, "dislike"},
		"GetPreferenceOfArtist: Likes gets the like tag":               {"/intern/preference/1/Sum%2041", internal, 11, http.StatusOK, "like"},
		"GetPreferenceOfArtist: Dislikes gets the dislike tag":         {"/intern/preference/1/ZZ%20Top", internal, 21, http.StatusOK, "dislike"},
		"GetPreferenceOfArtist: Non-preference songs are excluded":     {"/intern/preference/1/Sum%2041", internal, 404, http.StatusOK, ""},
		"GetPreferenceOfArtist: Songs from other artists are excluded": {"/intern/preference/1/Disturbed", internal, 11, http.StatusOK, ""},
	}
	for name, test := range cases {
		db := newTestDB()
		if err := db.AddUser(user); err != nil {
			t.Fatalf("Failed to start test due to failure of adding user %v: %s\n", user.Username, err)
		}
		db.addPreferencesToTestDB(t, user.ID, likedSongs, db.AddLike)
		db.addPreferencesToTestDB(t, user.ID, dislikedSongs, db.AddDislike)
		db.addSongsToTestDB(t, otherSongs)
		testServer := testServer(t, db, addDBToArray(make([]types.Song, 0), db))
		response := testhelpers.TestRequest(t, testServer, http.MethodGet, test.path, test.token, nil)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
		if response.Code != http.StatusOK {
			continue
		}
		var resultsMap map[int]string
		if err := convert.ReadFromJSONNoValidation(&resultsMap, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %s\n", name, err)
			continue
		}
		if resultsMap[test.idExpectedTag] != test.expectedTag {
			t.Errorf("%v: Expects song #%v with tag %v but got: %v\n", name, test.idExpectedTag, test.expectedTag, resultsMap[test.idExpectedTag])
		}
	}
}

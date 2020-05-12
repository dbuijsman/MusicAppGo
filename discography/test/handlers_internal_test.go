package test

import (
	"discography/handlers"
	"general/convert"
	"general/server"
	"general/testhelpers"
	"general/types"
	"net/http"
	"strconv"
	"testing"
)

func TestGetByID_response(t *testing.T) {
	var internal, user, admin string
	var err error
	if internal, err = server.CreateTokenInternalRequests("testServer"); err != nil {
		t.Fatalf("Failed to create internal token: %s\n", err)
	}
	if user, err = server.CreateToken(1, "test", "user"); err != nil {
		t.Fatalf("Failed to create user token: %s\n", err)
	}
	if admin, err = server.CreateToken(2, "testAdmin", "admin"); err != nil {
		t.Fatalf("Failed to create admin token: %s\n", err)
	}
	var artist types.Artist
	var song types.Song
	fakeID := -404
	clientSong := handlers.NewClientSong("Lost in Hollywood", "System of a Down")
	cases := map[string]struct {
		path, token        string
		id                 *int
		expectedStatusCode int
	}{
		"SongByID: Valid token and existing song":   {"/intern/song/", internal, &song.ID, http.StatusOK},
		"SongByID: User token is not authorized":    {"/intern/song/", user, &song.ID, http.StatusUnauthorized},
		"SongByID: Admin token is not authorized":   {"/intern/song/", admin, &song.ID, http.StatusUnauthorized},
		"SongByID: No token returns unauthorized":   {"/intern/song/", "", &song.ID, http.StatusUnauthorized},
		"SongByID: Song doesn't exist":              {"/intern/song/", internal, &fakeID, http.StatusNotFound},
		"ArtistByID: Valid token and existing song": {"/intern/artist/", internal, &artist.ID, http.StatusOK},
		"ArtistByID: User token is not authorized":  {"/intern/artist/", user, &artist.ID, http.StatusUnauthorized},
		"ArtistByID: Admin token is not authorized": {"/intern/artist/", admin, &artist.ID, http.StatusUnauthorized},
		"ArtistByID: No token returns unauthorized": {"/intern/artist/", "", &artist.ID, http.StatusUnauthorized},
		"ArtistByID: Song doesn't exist":            {"/intern/artist/", internal, &fakeID, http.StatusNotFound},
	}
	for name, test := range cases {
		db := newTestDB()
		if artist, err = db.AddArtist(clientSong.Artists[0], "", "link"); err != nil {
			t.Fatalf("Failed to start test due to failure of adding artist:%s\n", err)
		}
		if song, err = db.AddSong(clientSong.Name, []types.Artist{artist}); err != nil {
			t.Fatalf("Failed to start test due to failure of adding song:%s\n", err)
		}
		testServer, _ := testServerNoRequest(t, db)
		response := testhelpers.TestRequest(t, testServer, http.MethodGet, test.path+strconv.Itoa(*test.id), test.token, nil)
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
		if response.Code != http.StatusOK {
			continue
		}
		var result struct {
			ID int
		}
		if err := convert.ReadFromJSON(&result, response.Body); err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", name, err)
			continue
		}
		if result.ID != *test.id {
			t.Errorf("%v: Expects song with id %v but got: %v\n", name, *test.id, result.ID)
		}
	}
}

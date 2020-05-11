package test

import (
	"general"
	"testing"
)

func TestAddUser_saveInDB(t *testing.T) {
	cases := map[string]struct {
		id                int
		username, role    string
		expectedSavedInDB bool
	}{
		"No valid id":   {0, "Test", "user", false},
		"No username":   {1, "", "user", false},
		"No role":       {1, "TestNoRole", "", true},
		"Complete data": {1, "TestNoRole", "user", true},
	}
	for name, test := range cases {
		db := newTestDB()
		handler := testLikesHandler(db, nil)
		creds := general.NewCredentials(test.id, test.username, test.role)
		credsString, err := general.ToJSONBytes(creds)
		if err != nil {
			t.Errorf("%v: Can't serialize %v due to: %s\n", name, creds, err)
			continue
		}
		handler.ConsumeNewUser(credsString)
		if _, ok := db.users[test.id]; ok != test.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", name, test.expectedSavedInDB, ok)
		}
	}
}

func TestAddArtist_saveInDB(t *testing.T) {
	cases := map[string]struct {
		id                int
		name, prefix      string
		expectedSavedInDB bool
	}{
		"No valid id":   {0, "Rolling Stones", "The", false},
		"No name":       {1, "", "The", false},
		"No prefix":     {1, "Sum 41", "", true},
		"Complete data": {1, "Strokes", "The", true},
	}
	for name, test := range cases {
		db := newTestDB()
		handler := testLikesHandler(db, nil)
		artist := general.NewArtist(test.id, test.name, test.prefix)
		artistString, err := general.ToJSONBytes(artist)
		if err != nil {
			t.Errorf("%v: Can't serialize %v due to: %s\n", name, artist, err)
			continue
		}
		handler.ConsumeNewArtist(artistString)
		if _, ok := db.artists[test.name]; ok != test.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", name, test.expectedSavedInDB, ok)
		}
	}
}

func TestAddSong_saveInDB(t *testing.T) {
	existingArtist := general.NewArtist(1, "Lost Frequencies", "")
	cases := map[string]struct {
		id                int
		artists           []general.Artist
		name              string
		expectedSavedInDB bool
	}{
		"No valid id":                        {0, []general.Artist{general.NewArtist(2, "Miike Snow", "")}, "Genghis Khan", false},
		"No artist":                          {1, []general.Artist{}, "Paint It Black", false},
		"No song":                            {1, []general.Artist{general.NewArtist(2, "Miike Snow", "")}, "", false},
		"Complete data with existing artist": {1, []general.Artist{existingArtist}, "Reality", true},
		"Complete data with new artist":      {1, []general.Artist{general.NewArtist(2, "Miike Snow", "")}, "Genghis Khan", true},
		"Complete data with new artist and existing artist": {1, []general.Artist{general.NewArtist(2, "Zonderling", ""), existingArtist}, "Crazy", true},
	}
	for name, test := range cases {
		db := newTestDB()
		handler := testLikesHandler(db, nil)
		existingArtistString, err := general.ToJSONBytes(existingArtist)
		handler.ConsumeNewArtist(existingArtistString)
		if err != nil {
			t.Errorf("%v: Can't serialize %v due to: %s\n", name, existingArtist, err)
			continue
		}
		song := general.NewSong(test.id, test.artists, test.name)
		songString, err := general.ToJSONBytes(song)
		if err != nil {
			t.Errorf("%v: Can't serialize %v due to: %s\n", name, song, err)
			continue
		}
		handler.ConsumeNewSong(songString)
		if _, ok := db.songs[test.id]; ok != test.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", name, test.expectedSavedInDB, ok)
		}
	}
}

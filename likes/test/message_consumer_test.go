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
	for nameCase, caseCreds := range cases {
		db := newTestDB()
		handler := testLikesHandlerNilRequest(db)
		creds := general.NewCredentials(caseCreds.id, caseCreds.username, caseCreds.role)
		credsString, err := general.ToJSONBytes(creds)
		if err != nil {
			t.Fatalf("%v: Can't serialize %v due to: %s\n", nameCase, creds, err)
		}
		handler.ConsumeNewUser(credsString)
		if _, ok := db.users[caseCreds.id]; ok != caseCreds.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", nameCase, caseCreds.expectedSavedInDB, ok)
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
	for nameCase, caseArtist := range cases {
		db := newTestDB()
		handler := testLikesHandlerNilRequest(db)
		artist := general.NewArtist(caseArtist.id, caseArtist.name, caseArtist.prefix)
		artistString, err := general.ToJSONBytes(artist)
		if err != nil {
			t.Fatalf("%v: Can't serialize %v due to: %s\n", nameCase, artist, err)
		}
		handler.ConsumeNewArtist(artistString)
		if _, ok := db.artists[caseArtist.name]; ok != caseArtist.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", nameCase, caseArtist.expectedSavedInDB, ok)
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
	for nameCase, caseSong := range cases {
		db := newTestDB()
		handler := testLikesHandlerNilRequest(db)
		existingArtistString, err := general.ToJSONBytes(existingArtist)
		handler.ConsumeNewArtist(existingArtistString)
		if err != nil {
			t.Fatalf("%v: Can't serialize %v due to: %s\n", nameCase, existingArtist, err)
		}
		song := general.NewSong(caseSong.id, caseSong.artists, caseSong.name)
		songString, err := general.ToJSONBytes(song)
		if err != nil {
			t.Fatalf("%v: Can't serialize %v due to: %s\n", nameCase, song, err)
		}
		handler.ConsumeNewSong(songString)
		if _, ok := db.songs[caseSong.id]; ok != caseSong.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved: %v but got %v\n", nameCase, caseSong.expectedSavedInDB, ok)
		}
	}
}

package test

import (
	"discography/database"
	"discography/handlers"
	"general"
	"math"
	"net/http"
	"sort"
	"strconv"
	"testing"
)

func testServerNoRequest(t *testing.T, db database.Database) (*http.Server, chan general.Message) {
	handler, channel := testMusicHandlerNoRequest(t, db)
	server, _ := handlers.NewMusicServer(handler, nil, "music_test", "")
	return server, channel
}

func testMusicHandlerNoRequest(t *testing.T, db database.Database) (*handlers.MusicHandler, chan general.Message) {
	logger := general.TestEmptyLogger()
	sendMessage, channel := general.TestSendMessage()
	get := func(url string) (*http.Response, error) {
		response := http.Response{StatusCode: http.StatusNotImplemented}
		return &response, nil
	}
	handler, err := handlers.NewMusicHandler(logger, db, sendMessage, get)
	if err != nil {
		t.Fatalf("Failed to create a testServer due to: %s\n", err)
	}
	return handler, channel
}

func testAddDiscographyToDB(t *testing.T, db database.Database, discography map[string][]handlers.ClientSong) error {
	handler, _ := testMusicHandlerNoRequest(t, db)
	for _, discographyArtist := range discography {
		for _, song := range discographyArtist {
			if _, err := handler.AddSong(song.Name, song.Artists...); err != nil {
				return err
			}
		}
	}
	return nil
}

type testDB struct {
	artistsDB map[string]testArtist
	songsDB   map[string]map[string]general.Song
	lastID    int
}

func newTestDB() testDB {
	return testDB{artistsDB: make(map[string]testArtist), songsDB: make(map[string]map[string]general.Song)}
}

type testArtist struct {
	id                        int
	name, prefix, linkSpotify string
}

func (fake testDB) GetArtistsStartingWithLetter(startLetter string, offset, max int) ([]general.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.artistsDB {
		if artist[0:len(startLetter)] == startLetter {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	searchResults := make([]general.Artist, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		foundArtist := fake.artistsDB[allResults[indexResult]]
		searchResults = append(searchResults, general.NewArtist(foundArtist.id, foundArtist.name, foundArtist.prefix))
	}
	return searchResults, nil
}
func (fake testDB) GetArtistsStartingWithNumber(offset, max int) ([]general.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.artistsDB {
		if _, err := strconv.Atoi(artist[0:1]); err == nil {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	searchResults := make([]general.Artist, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		foundArtist := fake.artistsDB[allResults[indexResult]]
		searchResults = append(searchResults, general.NewArtist(foundArtist.id, foundArtist.name, foundArtist.prefix))
	}
	return searchResults, nil

}

func (fake testDB) GetSongsFromArtist(artist string, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	discography := fake.songsDB[artist]
	songNames := make([]string, 0, len(discography))
	for song := range discography {
		songNames = append(songNames, song)
	}
	sort.Strings(songNames)
	searchResults := make([]general.Song, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(songNames)))); indexResult++ {
		searchResults = append(searchResults, discography[songNames[indexResult]])

	}
	return searchResults, nil
}

func (fake testDB) FindArtistByName(name string) (general.Artist, error) {
	artist, ok := fake.artistsDB[name]
	if !ok {
		return general.Artist{}, general.GetDBError("Not found", general.NotFoundError)
	}
	return general.NewArtist(artist.id, artist.name, artist.prefix), nil
}

func (fake testDB) FindSongByName(artist, song string) (general.Song, error) {
	discography, ok := fake.songsDB[artist]
	if !ok {
		return general.Song{}, general.GetDBError("Not found", general.NotFoundError)
	}
	songData, okSong := discography[song]
	if !okSong {
		return general.Song{}, general.GetDBError("Not found", general.NotFoundError)
	}
	return songData, nil
}

func (fake testDB) FindArtistByID(artistID int) (general.Artist, error) {
	for _, artist := range fake.artistsDB {
		if artist.id == artistID {
			return general.NewArtist(artist.id, artist.name, artist.prefix), nil
		}
	}
	return general.Artist{}, general.GetDBError("Not found", general.NotFoundError)
}

func (fake testDB) FindSongByID(songID int) (general.Song, error) {
	for _, discography := range fake.songsDB {
		for _, song := range discography {
			if song.ID == songID {
				return song, nil
			}
		}
	}
	return general.Song{}, general.GetDBError("Not found", general.NotFoundError)
}

func (fake testDB) AddArtist(artist, prefix, linkSpotify string) (general.Artist, error) {
	if len(artist) == 0 {
		return general.Artist{}, general.GetDBError("Missing name", general.InvalidInput)
	}
	if _, ok := fake.artistsDB[artist]; ok {
		return general.Artist{}, general.GetDBError("Duplicate artist", general.DuplicateEntry)
	}
	newArtist := testArtist{id: len(fake.artistsDB) + 1, name: artist, prefix: prefix, linkSpotify: linkSpotify}
	fake.artistsDB[artist] = newArtist
	fake.songsDB[artist] = make(map[string]general.Song)
	return general.NewArtist(newArtist.id, newArtist.name, newArtist.prefix), nil
}

func (fake testDB) AddSong(song string, artists []general.Artist) (general.Song, error) {
	if len(song) == 0 {
		return general.Song{}, general.GetDBError("Missing name", general.InvalidInput)
	}
	if len(artists) == 0 {
		return general.Song{}, general.GetDBError("No artists is given for adding a song", general.InvalidInput)
	}
	for _, artist := range artists {
		if _, ok := fake.songsDB[artist.Name]; !ok {
			return general.Song{}, general.GetDBError("Artist doesn't exist", general.UnknownError)
		}
		if _, ok := fake.songsDB[artist.Name][song]; ok {
			return general.Song{}, general.GetDBError("Duplicate entry", general.DuplicateEntry)
		}
	}
	songID := artists[0].ID*10 + len(fake.songsDB[artists[0].Name])
	newSong := general.Song{ID: songID, Name: song, Artists: artists}
	for _, artist := range artists {
		fake.songsDB[artist.Name][song] = newSong
	}
	return newSong, nil
}

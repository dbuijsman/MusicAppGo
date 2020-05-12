package test

import (
	"discography/database"
	"discography/handlers"
	"general/dberror"
	"general/testhelpers"
	"general/types"
	"math"
	"net/http"
	"sort"
	"strconv"
	"testing"
)

func testServerNoRequest(t *testing.T, db database.Database) (*http.Server, chan testhelpers.Message) {
	handler, channel := testMusicHandlerNoRequest(t, db)
	testServer, _ := handlers.NewMusicServer(handler, nil, "music_test", "")
	return testServer, channel
}

func testMusicHandlerNoRequest(t *testing.T, db database.Database) (*handlers.MusicHandler, chan testhelpers.Message) {
	logger := testhelpers.TestEmptyLogger()
	sendMessage, channel := testhelpers.TestSendMessage()
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
	songsDB   map[string]map[string]types.Song
	lastID    int
}

func newTestDB() testDB {
	return testDB{artistsDB: make(map[string]testArtist), songsDB: make(map[string]map[string]types.Song)}
}

type testArtist struct {
	id                        int
	name, prefix, linkSpotify string
}

func (fake testDB) GetArtistsStartingWithLetter(startLetter string, offset, max int) ([]types.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.artistsDB {
		if artist[0:len(startLetter)] == startLetter {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	searchResults := make([]types.Artist, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		foundArtist := fake.artistsDB[allResults[indexResult]]
		searchResults = append(searchResults, types.NewArtist(foundArtist.id, foundArtist.name, foundArtist.prefix))
	}
	return searchResults, nil
}
func (fake testDB) GetArtistsStartingWithNumber(offset, max int) ([]types.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.artistsDB {
		if _, err := strconv.Atoi(artist[0:1]); err == nil {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	searchResults := make([]types.Artist, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		foundArtist := fake.artistsDB[allResults[indexResult]]
		searchResults = append(searchResults, types.NewArtist(foundArtist.id, foundArtist.name, foundArtist.prefix))
	}
	return searchResults, nil

}

func (fake testDB) GetSongsFromArtist(artist string, offset, max int) ([]types.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	discography := fake.songsDB[artist]
	songNames := make([]string, 0, len(discography))
	for song := range discography {
		songNames = append(songNames, song)
	}
	sort.Strings(songNames)
	searchResults := make([]types.Song, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(songNames)))); indexResult++ {
		searchResults = append(searchResults, discography[songNames[indexResult]])

	}
	return searchResults, nil
}

func (fake testDB) FindArtistByName(name string) (types.Artist, error) {
	artist, ok := fake.artistsDB[name]
	if !ok {
		return types.Artist{}, dberror.GetDBError("Not found", dberror.NotFoundError)
	}
	return types.NewArtist(artist.id, artist.name, artist.prefix), nil
}

func (fake testDB) FindSongByName(artist, song string) (types.Song, error) {
	discography, ok := fake.songsDB[artist]
	if !ok {
		return types.Song{}, dberror.GetDBError("Not found", dberror.NotFoundError)
	}
	songData, okSong := discography[song]
	if !okSong {
		return types.Song{}, dberror.GetDBError("Not found", dberror.NotFoundError)
	}
	return songData, nil
}

func (fake testDB) FindArtistByID(artistID int) (types.Artist, error) {
	for _, artist := range fake.artistsDB {
		if artist.id == artistID {
			return types.NewArtist(artist.id, artist.name, artist.prefix), nil
		}
	}
	return types.Artist{}, dberror.GetDBError("Not found", dberror.NotFoundError)
}

func (fake testDB) FindSongByID(songID int) (types.Song, error) {
	for _, discography := range fake.songsDB {
		for _, song := range discography {
			if song.ID == songID {
				return song, nil
			}
		}
	}
	return types.Song{}, dberror.GetDBError("Not found", dberror.NotFoundError)
}

func (fake testDB) AddArtist(artist, prefix, linkSpotify string) (types.Artist, error) {
	if len(artist) == 0 {
		return types.Artist{}, dberror.GetDBError("Missing name", dberror.InvalidInput)
	}
	if _, ok := fake.artistsDB[artist]; ok {
		return types.Artist{}, dberror.GetDBError("Duplicate artist", dberror.DuplicateEntry)
	}
	newArtist := testArtist{id: len(fake.artistsDB) + 1, name: artist, prefix: prefix, linkSpotify: linkSpotify}
	fake.artistsDB[artist] = newArtist
	fake.songsDB[artist] = make(map[string]types.Song)
	return types.NewArtist(newArtist.id, newArtist.name, newArtist.prefix), nil
}

func (fake testDB) AddSong(song string, artists []types.Artist) (types.Song, error) {
	if len(song) == 0 {
		return types.Song{}, dberror.GetDBError("Missing name", dberror.InvalidInput)
	}
	if len(artists) == 0 {
		return types.Song{}, dberror.GetDBError("No artists is given for adding a song", dberror.InvalidInput)
	}
	for _, artist := range artists {
		if _, ok := fake.songsDB[artist.Name]; !ok {
			return types.Song{}, dberror.GetDBError("Artist doesn't exist", dberror.UnknownError)
		}
		if _, ok := fake.songsDB[artist.Name][song]; ok {
			return types.Song{}, dberror.GetDBError("Duplicate entry", dberror.DuplicateEntry)
		}
	}
	songID := artists[0].ID*10 + len(fake.songsDB[artists[0].Name])
	newSong := types.Song{ID: songID, Name: song, Artists: artists}
	for _, artist := range artists {
		fake.songsDB[artist.Name][song] = newSong
	}
	return newSong, nil
}

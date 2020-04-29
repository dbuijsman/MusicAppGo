package test

import (
	"discography/handlers"
	"general"
	"math"
	"sort"
	"strconv"
)

func testMusicHandler() *handlers.MusicHandler {
	l := general.TestEmptyLogger()
	return handlers.NewMusicHandler(l, newTestDB(), general.TestSendMessageEmpty(), nil)
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
	if _, ok := fake.artistsDB[artist]; ok {
		return general.Artist{}, general.GetDBError("Duplicate artist", general.DuplicateEntry)
	}
	newArtist := testArtist{id: len(fake.artistsDB) + 1, name: artist, prefix: prefix, linkSpotify: linkSpotify}
	fake.artistsDB[artist] = newArtist
	fake.songsDB[artist] = make(map[string]general.Song)
	return general.NewArtist(newArtist.id, newArtist.name, newArtist.prefix), nil
}

func (fake testDB) AddSong(song string, artists []general.Artist) (general.Song, error) {
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

type song struct {
	artists  []string
	nameSong string
}

func getNewSong(nameSong string, artists ...string) song {
	return song{artists: artists, nameSong: nameSong}
}

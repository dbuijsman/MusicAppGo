package test

import (
	"MusicAppGo/common"
	"discography/database"
	"discography/handlers"
	"log"
	"math"
	"sort"
)

func testMusicHandler() *handlers.MusicHandler {
	l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
	handler := handlers.NewMusicHandler(l, newTestDB())
	return handler
}

type testWriter struct{}

func (fake testWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	return
}

type testDB struct {
	artistsDB map[string]database.RowArtistDB
	songsDB   map[string]map[string]database.SongDB
}

func newTestDB() testDB {
	return testDB{artistsDB: make(map[string]database.RowArtistDB), songsDB: make(map[string]map[string]database.SongDB)}
}

func (fake testDB) GetArtistsStartingWith(startLetter string, offset, max int) ([]database.RowArtistDB, error) {
	if max <= 0 || offset < 0 {
		return nil, common.GetDBError("Can not search with negative offset or non-positive max", common.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.artistsDB {
		if artist[0:len(startLetter)] == startLetter {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	searchResults := make([]database.RowArtistDB, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		searchResults = append(searchResults, fake.artistsDB[allResults[indexResult]])
	}
	return searchResults, nil
}

func (fake testDB) GetSongsFromArtist(artist string, offset, max int) ([]database.RowSongDB, error) {
	discography := fake.songsDB[artist]
	name_songs := make([]string, 0, len(discography))
	for song := range discography {
		name_songs = append(name_songs, song)
	}
	sort.Strings(name_songs)
	searchResults := make([]database.RowSongDB, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(name_songs)))); indexResult++ {
		song := discography[name_songs[indexResult]]
		for _, artist := range song.Artists {
			searchResults = append(searchResults, database.RowSongDB{ArtistName: artist.Artist, ArtistPrefix: artist.Prefix, SongName: song.Song})
		}
	}
	return searchResults, nil
}

func (fake testDB) FindArtist(name string) (database.RowArtistDB, error) {
	artist, ok := fake.artistsDB[name]
	if !ok {
		return database.RowArtistDB{}, common.GetDBError("Not found", common.NotFoundError)
	}
	return artist, nil
}
func (fake testDB) FindSong(artist, song string) (database.SongDB, error) {
	discography, ok := fake.songsDB[artist]
	if !ok {
		return database.SongDB{}, common.GetDBError("Not found", common.NotFoundError)
	}
	songData, okSong := discography[song]
	if !okSong {
		return database.SongDB{}, common.GetDBError("Not found", common.NotFoundError)
	}
	return songData, nil
}
func (fake testDB) AddArtist(artist, prefix, linkSpotify string) (database.RowArtistDB, error) {
	if _, ok := fake.artistsDB[artist]; ok {
		return database.RowArtistDB{}, common.GetDBError("Duplicate artist", common.DuplicateEntry)
	}
	newArtist := database.RowArtistDB{Artist: artist, Prefix: prefix, LinkSpotify: linkSpotify}
	fake.artistsDB[artist] = newArtist
	fake.songsDB[artist] = make(map[string]database.SongDB)
	return newArtist, nil
}

func (fake testDB) AddSong(song string, artists []database.RowArtistDB) (database.SongDB, error) {
	if len(artists) == 0 {
		return database.SongDB{}, common.GetDBError("No artists is given for adding a song", common.IncompleteInput)
	}
	for _, artist := range artists {
		if _, ok := fake.songsDB[artist.Artist]; !ok {
			return database.SongDB{}, common.GetDBError("Artist doesn't exist", common.UnknownError)
		}
	}
	newSong := database.SongDB{Song: song, Artists: artists}
	for _, artist := range artists {
		fake.songsDB[artist.Artist][song] = newSong
	}
	return newSong, nil
}

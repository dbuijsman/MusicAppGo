package test

import (
	"common"
	"discography/database"
	"discography/handlers"
	"log"
	"math"
	"sort"
)

func testMusicHandler() *handlers.MusicHandler {
	l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
	handler := handlers.NewMusicHandler(l, testDB{db: make(map[string]database.RowArtistDB)})
	return handler
}

type testWriter struct{}

func (fake testWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	return
}

type testDB struct {
	db map[string]database.RowArtistDB
}

func (fake testDB) GetArtistsStartingWith(startLetter string, offset, max int) ([]database.RowArtistDB, error) {
	if max <= 0 || offset < 0 {
		return nil, common.GetDBError("Can not search with negative offset or non-positive max", common.InvalidOffsetMax)
	}
	allResults := make([]string, 0, max)
	for artist := range fake.db {
		if artist[0:len(startLetter)] == startLetter {
			allResults = append(allResults, artist)
		}
	}
	sort.Strings(allResults)
	if offset >= len(allResults) {
		return make([]database.RowArtistDB, 0), nil
	}
	searchResults := make([]database.RowArtistDB, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(allResults)))); indexResult++ {
		searchResults = append(searchResults, fake.db[allResults[indexResult]])
	}
	return searchResults, nil
}
func (fake testDB) AddArtist(artist, prefix, linkSpotify string) error {
	if _, ok := fake.db[artist]; ok {
		return common.GetDBError("Duplicate artist", common.DuplicateEntry)
	}
	fake.db[artist] = database.RowArtistDB{Artist: artist, Prefix: prefix, LinkSpotify: linkSpotify}
	return nil
}

package database

import (
	"database/sql"
	"general/dberror"
	"general/types"
	"sync"
)

func scanArtists(results *sql.Rows) ([]types.Artist, error) {
	returningResults := make([]types.Artist, 0, 20)
	for results.Next() {
		var artist types.Artist
		err := results.Scan(&artist.ID, &artist.Name, &artist.Prefix)
		if err != nil {
			return nil, dberror.GetDBError(err.Error(), dberror.ScannerError)
		}
		returningResults = append(returningResults, artist)
	}
	return returningResults, nil
}

func scanSongs(results *sql.Rows) ([]types.Song, error) {
	rowFromDBChannel, songChannel := make(chan artistAndSong, 20), make(chan types.Song, 20)
	go connectArtistsAndSongs(rowFromDBChannel, songChannel)
	var wg sync.WaitGroup
	wg.Add(1)
	returningResults := make([]types.Song, 0, 20)
	go func() {
		for song := range songChannel {
			returningResults = append(returningResults, song)
		}
		wg.Done()
	}()
	for results.Next() {
		song := types.NewSong(0, nil, "")
		var artist types.Artist
		err := results.Scan(&artist.ID, &artist.Name, &artist.Prefix, &song.ID, &song.Name)
		if err != nil {
			defer close(rowFromDBChannel)
			return nil, dberror.GetDBError(err.Error(), dberror.ScannerError)
		}
		rowFromDBChannel <- artistAndSong{artist: artist, song: song}
	}
	close(rowFromDBChannel)
	wg.Wait()
	if len(returningResults) == 0 {
		return nil, dberror.GetDBError("Not found", dberror.NotFoundError)
	}
	return returningResults, nil
}

type artistAndSong struct {
	artist types.Artist
	song   types.Song
}

func connectArtistsAndSongs(input <-chan artistAndSong, output chan<- types.Song) {
	defer close(output)
	var once sync.Once
	var lastFoundSong types.Song
	for artistSong := range input {
		once.Do(func() { lastFoundSong = artistSong.song })
		if artistSong.song.ID != lastFoundSong.ID {
			output <- lastFoundSong
			lastFoundSong = artistSong.song
		}
		lastFoundSong.Artists = append(lastFoundSong.Artists, artistSong.artist)
	}
	output <- lastFoundSong
}

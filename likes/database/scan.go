package database

import (
	"database/sql"
	"general/dberror"
	"general/types"
	"sync"
)

func scanSongs(results *sql.Rows, preference string) ([]types.Song, error) {
	rowFromDBChannel, songChannel := make(chan artistAndSong, 20), make(chan types.Song, 20)
	go connectArtistsAndSongs(rowFromDBChannel, songChannel, preference)
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

func connectArtistsAndSongs(input <-chan artistAndSong, output chan<- types.Song, preference string) {
	defer close(output)
	var once sync.Once
	var lastFoundSong types.Song
	for artistSong := range input {
		once.Do(func() {
			lastFoundSong = artistSong.song
			lastFoundSong.Preference = preference
		})
		if artistSong.song.ID != lastFoundSong.ID {
			output <- lastFoundSong
			lastFoundSong = artistSong.song
			lastFoundSong.Preference = preference
		}
		lastFoundSong.Artists = append(lastFoundSong.Artists, artistSong.artist)
	}
	if lastFoundSong.ID == 0 {
		return
	}
	output <- lastFoundSong
}

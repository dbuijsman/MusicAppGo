package database

import (
	"database/sql"
	"general"
	"sync"
)

func scanArtists(results *sql.Rows) ([]general.Artist, error) {
	returningResults := make([]general.Artist, 0, 20)
	for results.Next() {
		var artist general.Artist
		err := results.Scan(&artist.ID, &artist.Name, &artist.Prefix)
		if err != nil {
			return nil, general.GetDBError(err.Error(), general.ScannerError)
		}
		returningResults = append(returningResults, artist)
	}
	return returningResults, nil
}

func scanSongs(results *sql.Rows) ([]general.Song, error) {
	rowFromDBChannel, songChannel := make(chan artistAndSong, 20), make(chan general.Song, 20)
	go connectArtistsAndSongs(rowFromDBChannel, songChannel)
	var wg sync.WaitGroup
	wg.Add(1)
	returningResults := make([]general.Song, 0, 20)
	go func() {
		for song := range songChannel {
			returningResults = append(returningResults, song)
		}
		wg.Done()
	}()
	for results.Next() {
		song := general.NewSong(0, nil, "")
		var artist general.Artist
		err := results.Scan(&artist.ID, &artist.Name, &artist.Prefix, &song.ID, &song.Name)
		if err != nil {
			defer close(rowFromDBChannel)
			return nil, general.GetDBError(err.Error(), general.ScannerError)
		}
		rowFromDBChannel <- artistAndSong{artist: artist, song: song}
	}
	close(rowFromDBChannel)
	wg.Wait()
	if len(returningResults) == 0 {
		return nil, general.GetDBError("Not found", general.NotFoundError)
	}
	return returningResults, nil
}

type artistAndSong struct {
	artist general.Artist
	song   general.Song
}

func connectArtistsAndSongs(input <-chan artistAndSong, output chan<- general.Song) {
	defer close(output)
	var once sync.Once
	var lastFoundSong general.Song
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

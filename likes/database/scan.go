package database

import (
	"database/sql"
	"general"
	"sync"
)

func scanSongs(results *sql.Rows, preference string) ([]general.Song, error) {
	rowFromDBChannel, songChannel := make(chan artistAndSong, 20), make(chan general.Song, 20)
	go connectArtistsAndSongs(rowFromDBChannel, songChannel, preference)
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

func connectArtistsAndSongs(input <-chan artistAndSong, output chan<- general.Song, preference string) {
	defer close(output)
	var once sync.Once
	var lastFoundSong general.Song
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

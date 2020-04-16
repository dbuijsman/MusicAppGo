package database

import (
	"MusicAppGo/common"
)

// AddArtist adds a new artist to the database
func (db *MusicDB) AddArtist(artist, prefix, linkSpotify string) error {
	_, err := db.database.Exec("INSERT INTO artists (name_artist, prefix, linkSpotify) VALUES ( ?, ?,?)", artist, prefix, linkSpotify)
	if err != nil {
		if err.Error()[6:10] == "1062" {
			return common.GetDBError(err.Error(), common.DuplicateEntry)
		}
		return common.GetDBError(err.Error(), common.UnknownError)
	}
	return nil
}

// AddSong will add a new song to the database
func (db *MusicDB) AddSong(song RowSongDB) error {
	for index, artist := range song.Artists {
		if artist.ID == 0 {
			newArtist, err := db.FindArtist(artist.Artist)
			if err != nil {
				return err
			}
			song.Artists[index] = newArtist
		}
	}

	return nil
}

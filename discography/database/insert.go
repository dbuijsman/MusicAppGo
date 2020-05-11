package database

import (
	"general"
)

// AddArtist adds a new artist to the database
func (db *MusicDB) AddArtist(artist, prefix, linkSpotify string) (general.Artist, error) {
	if len(artist) == 0 {
		return general.Artist{}, general.GetDBError("Missing name", general.InvalidInput)
	}
	resultID, err := db.database.Exec("INSERT INTO artists (name_artist, prefix, linkSpotify) VALUES ( ?, ?,?)", artist, prefix, linkSpotify)
	if err != nil {
		return general.Artist{}, general.MySQLErrorToDBError(err)
	}
	artistID, errorID := resultID.LastInsertId()
	if errorID != nil {
		return general.Artist{}, general.ErrorToUnknownDBError(errorID)
	}
	return general.NewArtist(int(artistID), artist, prefix), nil
}

// AddSong will add a new song to the database. This function won't check if the song already exists. It will return an error if the data is incomplete or if an artist don't exist.
func (db *MusicDB) AddSong(song string, artists []general.Artist) (general.Song, error) {
	if len(song) == 0 {
		return general.Song{}, general.GetDBError("Missing name", general.InvalidInput)
	}
	if len(artists) == 0 {
		return general.Song{}, general.GetDBError("No artists is given for adding a song", general.InvalidInput)
	}
	info, err := db.database.Exec("INSERT INTO songs (name_song) VALUES (?);", song)
	if err != nil {
		return general.Song{}, general.ErrorToUnknownDBError(err)
	}
	lastResult, errorID := info.LastInsertId()
	if errorID != nil {
		return general.Song{}, general.ErrorToUnknownDBError(errorID)
	}
	songID := int(lastResult)
	for _, artist := range artists {
		if artist.ID == 0 {
			return general.Song{}, general.GetDBError("Invalid ID for "+artist.Name, general.InvalidInput)
		}
		_, err = db.database.Exec("INSERT INTO discography (artist_id, song_id) VALUES (?,?);", artist.ID, songID)
		if err != nil {
			// Revert changes on failure
			db.database.Exec("DELETE FROM songs WHERE id=?;", songID)
			artistIDs := ""
			for _, artist := range artists {
				artistIDs += string(artist.ID) + ","
			}
			artistIDs = artistIDs[:len(artistIDs)-1]
			db.database.Exec("DELETE FROM discography WHERE song_id=? AND artist_id IN (?);", songID, artistIDs)
			return general.Song{}, general.MySQLErrorToDBError(err)
		}
	}
	return general.NewSong(songID, artists, song), nil
}

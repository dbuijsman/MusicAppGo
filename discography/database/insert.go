package database

import (
	"general/dberror"
	"general/types"
)

// AddArtist adds a new artist to the database
func (db *MusicDB) AddArtist(artist, prefix, linkSpotify string) (types.Artist, error) {
	if len(artist) == 0 {
		return types.Artist{}, dberror.GetDBError("Missing name", dberror.InvalidInput)
	}
	resultID, err := db.database.Exec("INSERT INTO artists (name_artist, prefix, linkSpotify) VALUES ( ?, ?,?)", artist, prefix, linkSpotify)
	if err != nil {
		return types.Artist{}, dberror.MySQLErrorToDBError(err)
	}
	artistID, errorID := resultID.LastInsertId()
	if errorID != nil {
		return types.Artist{}, dberror.ErrorToUnknownDBError(errorID)
	}
	return types.NewArtist(int(artistID), artist, prefix), nil
}

// AddSong will add a new song to the database. This function won't check if the song already exists. It will return an error if the data is incomplete or if an artist don't exist.
func (db *MusicDB) AddSong(song string, artists []types.Artist) (types.Song, error) {
	if len(song) == 0 {
		return types.Song{}, dberror.GetDBError("Missing name", dberror.InvalidInput)
	}
	if len(artists) == 0 {
		return types.Song{}, dberror.GetDBError("No artists is given for adding a song", dberror.InvalidInput)
	}
	info, err := db.database.Exec("INSERT INTO songs (name_song) VALUES (?);", song)
	if err != nil {
		return types.Song{}, dberror.ErrorToUnknownDBError(err)
	}
	lastResult, errorID := info.LastInsertId()
	if errorID != nil {
		return types.Song{}, dberror.ErrorToUnknownDBError(errorID)
	}
	songID := int(lastResult)
	for _, artist := range artists {
		if artist.ID == 0 {
			return types.Song{}, dberror.GetDBError("Invalid ID for "+artist.Name, dberror.InvalidInput)
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
			return types.Song{}, dberror.MySQLErrorToDBError(err)
		}
	}
	return types.NewSong(songID, artists, song), nil
}

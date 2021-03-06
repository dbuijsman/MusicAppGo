package database

import (
	"general"
)

// AddUser adds a new user to the database
func (db *LikesDB) AddUser(user general.Credentials) error {
	_, err := db.database.Exec("INSERT INTO users(id, username) VALUES (?,?)", user.ID, user.Username)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	return nil
}

// AddArtist adds a new artist to the database with the data from artist
func (db *LikesDB) AddArtist(artist general.Artist) error {
	resultID, err := db.database.Exec("INSERT INTO artists (id, name_artist, prefix) VALUES ( ?, ?,?)", artist.ID, artist.Name, artist.Prefix)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	_, errorID := resultID.LastInsertId()
	if errorID != nil {
		return general.ErrorToUnknownDBError(errorID)
	}
	return nil
}

// AddSong adds a new song to the database. It expects that
func (db *LikesDB) AddSong(song general.Song) error {
	if len(song.Artists) == 0 {
		return general.GetDBError("No artists is given for adding a song", general.InvalidInput)
	}
	_, err := db.database.Exec("INSERT INTO songs (id, name_song) VALUES (?,?);", song.ID, song.Name)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	for _, artist := range song.Artists {
		_, err = db.database.Exec("INSERT INTO discography (artist_id, song_id) VALUES (?,?);", artist.ID, song.ID)
		if err != nil {
			// Revert changes on failure
			db.database.Exec("DELETE FROM songs WHERE id=?;", song.ID)
			artistIDs := ""
			for _, artist := range song.Artists {
				artistIDs += string(artist.ID) + ","
			}
			artistIDs = artistIDs[:len(artistIDs)-1]
			db.database.Exec("DELETE FROM discography WHERE song_id=? AND artist_id IN (?);", song.ID, artistIDs)
			return general.MySQLErrorToDBError(err)
		}
	}
	return nil
}

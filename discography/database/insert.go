package database

import (
	"MusicAppGo/common"
)

// AddArtist adds a new artist to the database
func (db *MusicDB) AddArtist(artist, prefix, linkSpotify string) (RowArtistDB, error) {
	resultID, err := db.database.Exec("INSERT INTO artists (name_artist, prefix, linkSpotify) VALUES ( ?, ?,?)", artist, prefix, linkSpotify)
	if err != nil {
		if err.Error()[6:10] == "1062" {
			return RowArtistDB{}, common.GetDBError(err.Error(), common.DuplicateEntry)
		}
		return RowArtistDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	artistID, errorID := resultID.LastInsertId()
	if errorID != nil {
		return RowArtistDB{}, common.GetDBError(errorID.Error(), common.UnknownError)
	}
	return RowArtistDB{ID: int(artistID), Artist: artist, Prefix: prefix}, nil
}

// AddSong will add a new song to the database. This function won't check if the song already exists. It will return an error if the data is incomplete or if an artist don't exist.
func (db *MusicDB) AddSong(song string, artists []RowArtistDB) (SongDB, error) {
	if len(artists) == 0 {
		return SongDB{}, common.GetDBError("No artists is given for adding a song", common.IncompleteInput)
	}
	info, err := db.database.Exec("INSERT INTO songs (name_song) VALUES (?);", song)
	if err != nil {
		return SongDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	lastResult, errorID := info.LastInsertId()
	if errorID != nil {
		return SongDB{}, common.GetDBError(errorID.Error(), common.UnknownError)
	}
	songID := int(lastResult)
	for _, artist := range artists {
		if artist.ID == 0 {
			return SongDB{}, common.GetDBError("Invalid ID for "+artist.Artist, common.IncompleteInput)
		}
		_, err = db.database.Exec("INSERT INTO discography (artist_id, song_id) VALUES (?,?);")
		if err != nil { // Revert changes on failure
			db.database.Exec("DELETE FROM songs WHERE id=?;", songID)
			artistIDs := ""
			for _, artist := range artists {
				artistIDs += string(artist.ID) + ","
			}
			artistIDs = artistIDs[:len(artistIDs)-1]
			db.database.Exec("DELETE FROM discography WHERE song_id=? AND artist_id IN (?);", songID, artistIDs)
			return SongDB{}, err
		}
	}
	return SongDB{ID: songID, Song: song, Artists: artists}, nil
}

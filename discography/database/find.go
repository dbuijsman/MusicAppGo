package database

import (
	"MusicAppGo/common"
	"database/sql"
)

// GetArtistsStartingWith finds all artists that starts with a certain string
func (db *MusicDB) GetArtistsStartingWith(startLetter string, offset, max int) ([]RowArtistDB, error) {
	if max <= 0 || offset < 0 {
		return nil, common.GetDBError("Can not search with negative offset or non-positive max", common.InvalidOffsetMax)
	}
	startLetter = startLetter + "%"
	results, err := db.database.Query("SELECT id, name_artist, prefix FROM artists WHERE name_artist LIKE ? ORDER BY name_artist LIMIT ?,?;", startLetter, offset, max)
	if err != nil {
		return nil, common.GetDBError(err.Error(), common.UnknownError)
	}
	returningResults := make([]RowArtistDB, 0, max)
	for results.Next() {
		var artist RowArtistDB
		err = results.Scan(&artist.ID, &artist.Artist, &artist.Prefix)
		if err != nil {
			return nil, common.GetDBError(err.Error(), common.ScannerError)
		}
		returningResults = append(returningResults, artist)
	}
	return returningResults, nil
}

// FindArtist searches the database for the artist. This function expects a name without prefix.
func (db *MusicDB) FindArtist(name string) (RowArtistDB, error) {
	result := db.database.QueryRow("SELECT id, name_artist, prefix FROM artists WHERE name_artist=? LIMIT 1;", name)
	var artist RowArtistDB
	err := result.Scan(&artist.ID, &artist.Artist, &artist.Prefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return RowArtistDB{}, common.GetDBError(err.Error(), common.NotFoundError)
		}
		return RowArtistDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	return artist, nil
}

// FindSong searches the database for a particular song. This function expects a name without prefix
func (db *MusicDB) FindSong(artist, song string) (SongDB, error) {
	result := db.database.QueryRow("SELECT songs.id FROM artists, discography, songs WHERE name_artist=? AND artist_id=artists.id AND songs.id=song_id AND name_song=? LIMIT 1;", artist, song)
	var songID int
	if err := result.Scan(&songID); err != nil {
		if err != sql.ErrNoRows {
			return SongDB{}, common.GetDBError(err.Error(), common.NotFoundError)
		}
		return SongDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	results, err := db.database.Query("SELECT artists.id, name_artist, prefix, name_song FROM artists, discography, songs WHERE artists.id=artist_id AND songs.id=song_id AND songs.id=?;", songID)
	if err != nil {
		return SongDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	resultsArtist := make([]RowArtistDB, 0)
	var nameSong string
	for results.Next() {
		var contributingArtist RowSongDB
		err = result.Scan(&contributingArtist.ArtistID, &contributingArtist.ArtistName, &contributingArtist.ArtistPrefix, &contributingArtist.SongName)
		if err != nil {
			return SongDB{}, common.GetDBError(err.Error(), common.ScannerError)
		}
		nameSong = contributingArtist.SongName
		resultsArtist = append(resultsArtist, RowArtistDB{ID: contributingArtist.ArtistID, Artist: contributingArtist.ArtistName, Prefix: contributingArtist.ArtistPrefix})
	}
	return SongDB{ID: songID, Artists: resultsArtist, Song: nameSong}, nil
}

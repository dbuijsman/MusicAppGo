package database

import (
	"common"
)

// GetArtistsStartingWith finds all artists that starts with a certain string
func (db *MusicDB) GetArtistsStartingWith(startLetter string, offset, max int) ([]RowArtistDB, error) {
	if max <= 0 || offset < 0 {
		return nil, common.GetDBError("Can not search with negative offset or non-positive max", common.InvalidOffsetMax)
	}
	startLetter = startLetter + "%"
	result, err := db.database.Query("SELECT id, name_artist, prefix FROM artists WHERE name_artist LIKE ? ORDER BY name_artist LIMIT ?,?;", startLetter, offset, max)
	if err != nil {
		return nil, common.GetDBError(err.Error(), common.UnknownError)
	}
	allResults := make([]RowArtistDB, 0, max)
	for result.Next() {
		var artist RowArtistDB
		err = result.Scan(&artist.ID, &artist.Artist, &artist.Prefix)
		if err != nil {
			return nil, common.GetDBError(err.Error(), common.ScannerError)
		}
		allResults = append(allResults, artist)
	}
	return allResults, nil
}

// FindArtist searches the database for the artist. This function expects a name without prefix.
func (db *MusicDB) FindArtist(name string) (RowArtistDB, error) {
	result := db.database.QueryRow("SELECT id, name_artist, prefix FROM artists WHERE name_artist=? LIMIT 1;", name)
	var artist RowArtistDB
	err := result.Scan(&artist.ID, &artist.Artist, &artist.Prefix)
	if err != nil {
		return RowArtistDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	return artist, nil
}

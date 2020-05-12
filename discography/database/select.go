package database

import (
	"database/sql"
	"general/dberror"
	"general/types"
)

// GetArtistsStartingWithLetter finds all artists that starts with a certain string
func (db *MusicDB) GetArtistsStartingWithLetter(startLetter string, offset, max int) ([]types.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	startLetter = startLetter + "%"
	results, err := db.database.Query("SELECT id, name_artist, prefix FROM artists WHERE name_artist LIKE ? ORDER BY name_artist LIMIT ?,?;", startLetter, offset, max)
	if err != nil {
		return nil, dberror.ErrorToUnknownDBError(err)
	}
	defer results.Close()
	return scanArtists(results)
}

// GetArtistsStartingWithNumber finds all artists that start with a number
func (db *MusicDB) GetArtistsStartingWithNumber(offset, max int) ([]types.Artist, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	results, err := db.database.Query("SELECT id, name_artist, prefix FROM artists WHERE name_artist REGEXP '^(0|1|2|3|4|5|6|7|8|9)' ORDER BY name_artist LIMIT ?,?;", offset, max)
	if err != nil {
		return nil, dberror.ErrorToUnknownDBError(err)
	}
	defer results.Close()
	return scanArtists(results)

}

//GetSongsFromArtist finds songs of the given artist ordered by name of the song. The results are not yet combined (i.e. if multiple artists contributed on one song).
func (db *MusicDB) GetSongsFromArtist(artist string, offset, max int) ([]types.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, dberror.GetDBError("Can not search with negative offset or non-positive max", dberror.InvalidOffsetMax)
	}
	// This query is a cross join between artists, discography and a subquery that selects the songs of an artist
	results, err := db.database.Query("SELECT artists.id, name_artist, prefix, songsOfArtist.song_id, name_song FROM artists, discography CROSS JOIN (SELECT song_id, name_song FROM artists, discography, songs WHERE name_artist=? AND songs.id=song_id AND artists.id=artist_id ORDER BY name_song LIMIT ?,?) AS songsOfArtist ON discography.song_id=songsOfArtist.song_id WHERE artists.id=artist_id AND songsOfArtist.song_id=discography.song_id ORDER BY name_song;", artist, offset, max)
	if err != nil {
		return nil, dberror.ErrorToUnknownDBError(err)
	}
	defer results.Close()
	return scanSongs(results)
}

// FindArtistByName searches the database for the artist. This function expects a name without prefix.
// This function will be used for updating the DB (e.g. add a song of the given artist)
func (db *MusicDB) FindArtistByName(name string) (types.Artist, error) {
	result := db.database.QueryRow("SELECT id, name_artist, prefix FROM artists WHERE name_artist=? LIMIT 1;", name)
	var artist types.Artist
	err := result.Scan(&artist.ID, &artist.Name, &artist.Prefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Artist{}, dberror.GetDBError(err.Error(), dberror.NotFoundError)
		}
		return types.Artist{}, dberror.ErrorToUnknownDBError(err)
	}
	return artist, nil
}

// FindSongByName searches the database for a particular song. This function expects a name without prefix
// This function will be used for updating the DB (e.g. add an album with the given song)
func (db *MusicDB) FindSongByName(artist, song string) (types.Song, error) {
	result := db.database.QueryRow("SELECT songs.id FROM artists, discography, songs WHERE name_artist=? AND artist_id=artists.id AND songs.id=song_id AND name_song=? LIMIT 1;", artist, song)
	var songID int
	if err := result.Scan(&songID); err != nil {
		if err != sql.ErrNoRows {
			return types.Song{}, dberror.GetDBError(err.Error(), dberror.NotFoundError)
		}
		return types.Song{}, dberror.ErrorToUnknownDBError(err)
	}
	return db.FindSongByID(songID)
}

// FindArtistByID returns the artist that belongs to the given ID
func (db *MusicDB) FindArtistByID(artistID int) (types.Artist, error) {
	result := db.database.QueryRow("SELECT id, name_artist, prefix FROM artists WHERE id=? LIMIT 1;", artistID)
	var artist types.Artist
	err := result.Scan(&artist.ID, &artist.Name, &artist.Prefix)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Artist{}, dberror.GetDBError(err.Error(), dberror.NotFoundError)
		}
		return types.Artist{}, dberror.ErrorToUnknownDBError(err)
	}
	return artist, nil
}

// FindSongByID returns the song that belongs to the given ID
func (db *MusicDB) FindSongByID(songID int) (types.Song, error) {
	results, err := db.database.Query("SELECT artists.id, name_artist, prefix, songs.id, name_song FROM artists, discography, songs WHERE artists.id=artist_id AND songs.id=song_id AND songs.id=?;", songID)
	if err != nil {
		return types.Song{}, dberror.ErrorToUnknownDBError(err)
	}
	defer results.Close()
	songs, scanError := scanSongs(results)
	if scanError != nil {
		return types.Song{}, scanError
	}
	return songs[0], nil
}

package database

import (
	"general"
	"log"
	"sync"
)

//GetLikes finds liked songs of the given user ordered by name of the song.
func (db *LikesDB) GetLikes(userID, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	// This query is a cross join between artists, discography and a subquery that selects the likes from an user
	results, err := db.database.Query("SELECT artists.id, name_artist, prefix, discography.song_id, name_song FROM artists, discography CROSS JOIN (SELECT DISTINCT liked_songs.song_id, name_song FROM liked_songs, songs WHERE user_id=? AND liked_songs.song_id=songs.id ORDER BY name_song LIMIT ?,?) AS likes ON discography.song_id=likes.song_id WHERE artists.id=artist_id ORDER BY name_artist, name_song;", userID, offset, max)
	if err != nil {
		return nil, general.ErrorToUnknownDBError(err)
	}
	return scanSongs(results, "like")
}

//GetDislikes finds disliked songs of the given user ordered by name of the song.
func (db *LikesDB) GetDislikes(userID, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	// This query is a cross join between artists, discography and a subquery that selects the dislikes from an user
	results, err := db.database.Query("SELECT artists.id, name_artist, prefix, discography.song_id, name_song FROM artists, discography CROSS JOIN (SELECT DISTINCT disliked_songs.song_id, name_song FROM disliked_songs, songs WHERE user_id=? AND disliked_songs.song_id=songs.id ORDER BY name_song LIMIT ?,?) AS dislikes ON discography.song_id=dislikes.song_id WHERE artists.id=artist_id ORDER BY name_artist, name_song;", userID, offset, max)
	if err != nil {
		return nil, general.ErrorToUnknownDBError(err)
	}
	return scanSongs(results, "dislike")
}

// GetLikesIDFromArtistName searches all the songIDs of songs of the given artist that are liked by the given user and sends these to the given channel
func (db *LikesDB) GetLikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup) {
	defer close(channel)
	results, err := db.database.Query("SELECT liked_songs.song_id FROM liked_songs, discography, artists WHERE user_id=? AND liked_songs.song_id=discography.song_id AND artist_id=artists.id AND name_artist=?;", userID, nameArtist)
	if err != nil {
		logger.Printf("[ERROR] Can't search db for likes of user #%v for artist %v due to: %v\n", userID, nameArtist, err)
		return
	}
	var songID int
	for results.Next() {
		err = results.Scan(&songID)
		if err != nil {
			logger.Printf("[ERROR] Scanerror while scanning song ids of likes of user #%v and artist %v due to %v\n", userID, nameArtist, err)
			return
		}
		channel <- songID
	}
	wg.Done()
}

// GetDislikesIDFromArtistName searches all the songIDs of songs of the given artist that are disliked by the given user and sends these to the given channel
func (db *LikesDB) GetDislikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup) {
	defer close(channel)
	results, err := db.database.Query("SELECT disliked_songs.song_id FROM disliked_songs, discography, artists WHERE user_id=? AND disliked_songs.song_id=discography.song_id AND artist_id=artists.id AND name_artist=?;", userID, nameArtist)
	if err != nil {
		logger.Printf("[ERROR] Can't search db for dislikes of user #%v for artist %v due to: %v\n", userID, nameArtist, err)
		return
	}
	var songID int
	for results.Next() {
		err = results.Scan(&songID)
		if err != nil {
			logger.Printf("[ERROR] Scanerror while scanning song ids of dislikes of user #%v and artist %v due to %v\n", userID, nameArtist, err)
			return
		}
		channel <- songID
	}
	wg.Done()
}

package database

import (
	"fmt"
	"general"
)

// AddLike adds a new like to the database
func (db *LikesDB) AddLike(userID, songID int) error {
	query := fmt.Sprintf("INSERT INTO liked_songs (user_id,song_id) SELECT * FROM (SELECT %v, %v) AS tmp WHERE NOT EXISTS ( SELECT user_id, song_id FROM liked_songs WHERE user_id=? AND song_id=?) LIMIT 1;", userID, songID)
	_, err := db.database.Exec(query, userID, songID)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	return nil
}

// AddDislike adds a new dislike to the database
func (db *LikesDB) AddDislike(userID, songID int) error {
	query := fmt.Sprintf("INSERT INTO disliked_songs (user_id,song_id) SELECT * FROM (SELECT %v, %v) AS tmp WHERE NOT EXISTS ( SELECT user_id, song_id FROM liked_songs WHERE user_id=? AND song_id=?) LIMIT 1;", userID, songID)
	_, err := db.database.Exec(query, userID, songID)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	return nil
}

// RemoveLike removes a like to the database
func (db *LikesDB) RemoveLike(userID, songID int) error {
	_, err := db.database.Exec("DELETE FROM liked_songs WHERE user_id=? AND song_id=?;", userID, songID)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	return nil
}

// RemoveDislike removes a dislike to the database
func (db *LikesDB) RemoveDislike(userID, songID int) error {
	_, err := db.database.Exec("DELETE FROM disliked_songs WHERE user_id=? AND song_id=?;", userID, songID)
	if err != nil {
		return general.MySQLErrorToDBError(err)
	}
	return nil
}

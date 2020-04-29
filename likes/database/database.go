package database

import (
	"database/sql"
	"general"
)

// Database is an interface for the likes database
type Database interface {
	AddUser(user general.Credentials) error
	AddArtist(artist general.Artist) error
	AddSong(song general.Song) error
	AddLike(userID, songID int) error
	AddDislike(userID, songID int) error
	RemoveLike(userID, songID int) error
	RemoveDislike(userID, songID int) error
}

// LikesDB is a database
type LikesDB struct {
	database *sql.DB
}

// NewLikesDB returns a MusicDB
func NewLikesDB(db *sql.DB) *LikesDB {
	return &LikesDB{database: db}
}

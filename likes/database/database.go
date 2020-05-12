package database

import (
	"database/sql"
	"general/types"
	"log"
	"sync"
)

// Database is an interface for the likes database
type Database interface {
	AddUser(user types.Credentials) error
	AddArtist(artist types.Artist) error
	AddSong(song types.Song) error
	AddLike(userID, songID int) error
	AddDislike(userID, songID int) error
	RemoveLike(userID, songID int) error
	RemoveDislike(userID, songID int) error
	GetLikes(userID, offset, max int) ([]types.Song, error)
	GetDislikes(userID, offset, max int) ([]types.Song, error)
	GetLikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup)
	GetDislikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup)
}

// LikesDB is a database
type LikesDB struct {
	database *sql.DB
}

// NewLikesDB returns a MusicDB
func NewLikesDB(db *sql.DB) *LikesDB {
	return &LikesDB{database: db}
}

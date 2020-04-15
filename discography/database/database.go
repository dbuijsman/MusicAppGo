package database

import (
	"database/sql"
)

// Database is an interface to go through the music database
type Database interface {
	GetArtistsStartingWith(startLetter string, offset, max int) ([]RowArtistDB, error)
	AddArtist(artist, prefix, linkSpotify string) error
}

// MusicDB is a database
type MusicDB struct {
	database *sql.DB
}

// NewMusicDB returns a MusicDB
func NewMusicDB(db *sql.DB) *MusicDB {
	return &MusicDB{database: db}
}

// RowArtistDB represents an artist from the database.
type RowArtistDB struct {
	ID          int    `json: "id"`
	Artist      string `json: "artist"`
	Prefix      string `json: "prefix"`
	LinkSpotify string `json: "-"`
}

// RowSongDB represents a song from the database with the contributing artists
type RowSongDB struct {
	ID      int           `json: "id"`
	Artists []RowArtistDB `json: "artists"`
}

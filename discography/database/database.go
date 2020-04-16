package database

import (
	"database/sql"
)

// Database is an interface to go through the music database
type Database interface {
	GetArtistsStartingWith(startLetter string, offset, max int) ([]RowArtistDB, error)
	FindArtist(name string) (RowArtistDB, error)
	FindSong(artist, song string) (SongDB, error)
	AddArtist(artist, prefix, linkSpotify string) (RowArtistDB, error)
	AddSong(song string, artists []RowArtistDB) (SongDB, error)
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

// SongDB represents a song from the database with the contributing artists
type SongDB struct {
	ID      int           `json: "id"`
	Artists []RowArtistDB `json: "artists"`
	Song    string        `json: "song"`
}

// RowSongDB is an entry from the database containing the song and an artist
type RowSongDB struct {
	ArtistID                 int
	ArtistName, ArtistPrefix string
	SongID                   int
	SongName                 string
}

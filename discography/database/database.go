package database

import (
	"database/sql"
	"general"
)

// Database is an interface to go through the music database
type Database interface {
	GetArtistsStartingWithLetter(startLetter string, offset, max int) ([]general.Artist, error)
	GetArtistsStartingWithNumber(offset, max int) ([]general.Artist, error)
	GetSongsFromArtist(artist string, offset, max int) ([]general.Song, error)
	FindArtistByName(name string) (general.Artist, error)
	FindSongByName(artist, song string) (general.Song, error)
	FindSongByID(songID int) (general.Song, error)
	AddArtist(artist, prefix, linkSpotify string) (general.Artist, error)
	AddSong(song string, artists []general.Artist) (general.Song, error)
}

// MusicDB is a database
type MusicDB struct {
	database *sql.DB
}

// NewMusicDB returns a MusicDB
func NewMusicDB(db *sql.DB) *MusicDB {
	return &MusicDB{database: db}
}

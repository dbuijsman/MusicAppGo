package database

import (
	"database/sql"
	"general/types"
)

// Database is an interface to go through the music database
type Database interface {
	GetArtistsStartingWithLetter(startLetter string, offset, max int) ([]types.Artist, error)
	GetArtistsStartingWithNumber(offset, max int) ([]types.Artist, error)
	GetSongsFromArtist(artist string, offset, max int) ([]types.Song, error)
	FindArtistByName(name string) (types.Artist, error)
	FindSongByName(artist, song string) (types.Song, error)
	FindArtistByID(artistID int) (types.Artist, error)
	FindSongByID(songID int) (types.Song, error)
	AddArtist(artist, prefix, linkSpotify string) (types.Artist, error)
	AddSong(song string, artists []types.Artist) (types.Song, error)
}

// MusicDB is a database
type MusicDB struct {
	database *sql.DB
}

// NewMusicDB returns a MusicDB
func NewMusicDB(db *sql.DB) *MusicDB {
	return &MusicDB{database: db}
}

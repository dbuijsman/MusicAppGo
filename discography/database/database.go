package database

import (
	"database/sql"
)

// Database is an interface to go through the music database
type Database interface {
	GetArtistsStartingWith(startLetter string, offset, max int) ([]RowArtistDB, error)
	GetSongsFromArtist(artist string, offset, max int) ([]RowSongDB, error)
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
	ID          int    `json:"id"`
	Artist      string `json:"artist"`
	Prefix      string `json:"prefix"`
	LinkSpotify string `json:"-"`
}

// NewRowArtistDB returns a RowArtistDB containing the given data
func NewRowArtistDB(id int, artist, prefix string) RowArtistDB {
	return RowArtistDB{ID: id, Artist: artist, Prefix: prefix}
}

// SongDB represents a song from the database with the contributing artists
type SongDB struct {
	ID      int           `json:"id"`
	Artists []RowArtistDB `json:"artists"`
	Song    string        `json:"song"`
}

// NewSongDB returns a SongDB containing the given data
func NewSongDB(id int, song string) SongDB {
	return SongDB{ID: id, Song: song, Artists: make([]RowArtistDB, 0, 1)}
}

// RowSongDB is an entry from the database containing the song and an artist
type RowSongDB struct {
	ArtistID                 int
	ArtistName, ArtistPrefix string
	SongID                   int
	SongName                 string
}

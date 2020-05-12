package types

// Service contains the name and address of a microservce. This type will be used for sharing this data with the services
type Service struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
}

// Credentials contains the credentials of an user
type Credentials struct {
	ID       int    `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
	Role     string `json:"role"`
}

// NewCredentials returns Credentials with the given data
func NewCredentials(id int, username, role string) Credentials {
	return Credentials{ID: id, Username: username, Role: role}
}

// Artist contains the id, name, prefix and a link about the artist
type Artist struct {
	ID     int    `json:"id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Prefix string `json:"prefix"`
}

// NewArtist returns an Artist with the given data
func NewArtist(id int, name, prefix string) Artist {
	return Artist{ID: id, Name: name, Prefix: prefix}
}

// Song contains the id, contributing artists and name of the song
type Song struct {
	ID         int      `json:"id" validate:"required"`
	Artists    []Artist `json:"artists" validate:"required"`
	Name       string   `json:"name" validate:"required"`
	Preference string   `json:"preference"`
}

// NewSong returns a Song containing the given data
func NewSong(id int, artists []Artist, song string) Song {
	if artists == nil {
		artists = make([]Artist, 0, 1)
	}
	return Song{ID: id, Name: song, Artists: artists}
}

// Preference represents a preference of an user with the id of the song or artist and the page where the request came from.
type Preference struct {
	ID   int    `json:"id" validate:"required"`
	Page string `json:"page"`
	//Page string `json:"page" validate:"required"`
}

// NewPreference returns a Preference with the given data
func NewPreference(id int, page string) Preference {
	return Preference{ID: id, Page: page}
}

// MultipleArtists represents the results of a request in a form containing the found artists and a boolean that shows if there are more results
type MultipleArtists struct {
	Data    []Artist `json:"music"`
	HasNext bool     `json:"hasNext"`
}

// MultipleSongs represents the results of a request in a form containing the found songs and a boolean that shows if there are more results
type MultipleSongs struct {
	Data    []Song `json:"music"`
	HasNext bool   `json:"hasNext"`
}

// Music represents the results of a request in a form containing the found artists and songs and a boolean that shows if there are more results
type Music struct {
	Data    []interface{} `json:"music"`
	HasNext bool          `json:"hasNext"`
}

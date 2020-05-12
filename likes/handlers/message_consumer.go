package handlers

import (
	"general/convert"
	"general/dberror"
	"general/server"
	"general/types"

	"github.com/optiopay/kafka/v2"
)

// StartConsuming will start all the consumers that belongs to the likes service
func (handler *LikesHandler) StartConsuming(broker *kafka.Broker) {
	go server.StartConsumer(broker, handler.Logger, "newUser", handler.ConsumeNewUser)
	go server.StartConsumer(broker, handler.Logger, "newArtist", handler.ConsumeNewArtist)
	go server.StartConsumer(broker, handler.Logger, "newSong", handler.ConsumeNewSong)
}

// ConsumeNewUser consumes a message and adds a new user to the database
func (handler *LikesHandler) ConsumeNewUser(message []byte) {
	var newUser types.Credentials
	if err := convert.FromJSONBytes(&newUser, message); err != nil {
		handler.Logger.Printf("Failed to deserialize message: %v due to: %s\n", string(message), err)
		return
	}
	if err := handler.db.AddUser(newUser); err != nil {
		if err.(dberror.DBError).ErrorCode != dberror.DuplicateEntry {
			handler.Logger.Printf("[ERROR] Failed to add new user %v to DB: %s\n", newUser.Username, err)
			return
		}
		handler.Logger.Printf("Adding user %v results in duplicate error.\n", newUser.Username)
	}
	handler.Logger.Printf("Succesfully added new user %v\n", newUser.Username)
}

// ConsumeNewArtist consumes a message and adds a new artist to the database
func (handler *LikesHandler) ConsumeNewArtist(message []byte) {
	var artist types.Artist
	if err := convert.FromJSONBytes(&artist, message); err != nil {
		handler.Logger.Printf("Failed to deserialize message: %v due to: %s\n", string(message), err)
		return
	}
	if err := handler.db.AddArtist(artist); err != nil {
		if err.(dberror.DBError).ErrorCode != dberror.DuplicateEntry {
			handler.Logger.Printf("[ERROR] Failed to add new artist %v to DB: %s\n", artist.Name, err)
			return
		}
		handler.Logger.Printf("Adding artist %v results in duplicate error.\n", artist.Name)
	}
	handler.Logger.Printf("Succesfully added new artist %v\n", artist.Name)
}

// ConsumeNewSong consumes a message and adds a new song to the database. It expects that the collaborating artists already exists
func (handler *LikesHandler) ConsumeNewSong(message []byte) {
	var song types.Song
	if err := convert.FromJSONBytes(&song, message); err != nil {
		handler.Logger.Printf("Failed to deserialize message: %v due to: %s\n", string(message), err)
		return
	}
	if len(song.Artists) == 0 {
		handler.Logger.Printf("Received new song #%v without artist\n", song.ID)
		return
	}
	if err := handler.db.AddSong(song); err != nil {
		switch err.(dberror.DBError).ErrorCode {
		case dberror.MissingForeignKey:
			for _, artist := range song.Artists {
				addArtistErr := handler.db.AddArtist(artist)
				if addArtistErr != nil && addArtistErr.(dberror.DBError).ErrorCode != dberror.DuplicateEntry {
					handler.Logger.Printf("Failed to add new song %v -%v due to failure of adding artist %v: %s\n", song.Artists[0].Name, song.Name, artist.Name, addArtistErr)
					return
				}
				if addArtistErr == nil {
					handler.Logger.Printf("Artist %v was missing from DB but is now succesfully added.\n", artist.Name)
				}
			}
			if errSecondTry := handler.db.AddSong(song); errSecondTry != nil {
				handler.Logger.Printf("Failed to add new song %v -%v due to second failure: %s\n", song.Artists[0].Name, song.Name, errSecondTry)
				return
			}
		case dberror.DuplicateEntry:
			handler.Logger.Printf("Adding song %v -%v results in duplicate error.\n", song.Artists[0].Name, song.Name)
			return
		default:
			handler.Logger.Printf("[ERROR] Failed to add new song %v - %v to DB: %s\n", song.Artists[0].Name, song.Name, err)
			return
		}
	}
	handler.Logger.Printf("Succesfully added new song %v -%v\n", song.Artists[0].Name, song.Name)
}

package test

import (
	"fmt"
	"general"
	"io"
	"likes/handlers"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func testLikesHandlerNilRequest(db testDB) *handlers.LikesHandler {
	l := general.TestEmptyLogger()
	return handlers.NewLikesHandler(l, db, nil)
}

func testLikesHandlerWithRequest(db testDB, existingSongs []general.Song) *handlers.LikesHandler {
	l := general.TestEmptyLogger()
	songDB := make(map[int]general.Song)
	for _, song := range existingSongs {
		songDB[song.ID] = song
	}
	get := func(address string) (*http.Response, error) {
		indexLastSlash := strings.LastIndex(address, "/")
		if indexLastSlash == -1 {
			return convertMessageInResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}
		songID, err := strconv.Atoi(address[indexLastSlash+1:])
		if err != nil {
			return convertMessageInResponse(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}
		song, ok := songDB[songID]
		if !ok {
			return convertMessageInResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}
		return convertMessageInResponse(http.StatusOK, song)
	}
	return handlers.NewLikesHandler(l, db, get)
}

func convertMessageInResponse(status int, body interface{}) (*http.Response, error) {
	var resp http.Response
	var err error
	resp.StatusCode = status
	bodyRequest, writer := io.Pipe()
	go func() {
		err = general.WriteToJSON(body, writer)
		writer.Close()
	}()
	resp.Body = bodyRequest
	if err != nil {
		return nil, fmt.Errorf("Error in test helper: %s", err)
	}
	return &resp, nil

}

type testDB struct {
	users    map[int]general.Credentials
	artists  map[string]general.Artist
	songs    map[int]general.Song
	likes    map[int]map[int]general.Song
	dislikes map[int]map[int]general.Song
}

func newTestDB() testDB {
	users := make(map[int]general.Credentials)
	artists := make(map[string]general.Artist)
	songs := make(map[int]general.Song)
	likes := make(map[int]map[int]general.Song)
	dislikes := make(map[int]map[int]general.Song)
	return testDB{users: users, artists: artists, songs: songs, likes: likes, dislikes: dislikes}
}

func (fake testDB) addSongsToTestDB(songs []general.Song) error {
	for _, song := range songs {
		err := fake.AddSong(song)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fake testDB) AddUser(user general.Credentials) error {
	if _, ok := fake.users[user.ID]; ok {
		return general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	fake.users[user.ID] = user
	fake.likes[user.ID] = make(map[int]general.Song)
	fake.dislikes[user.ID] = make(map[int]general.Song)
	return nil
}

func (fake testDB) AddArtist(artist general.Artist) error {
	if _, ok := fake.artists[artist.Name]; ok {
		return general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	fake.artists[artist.Name] = artist
	return nil
}

func (fake testDB) AddSong(song general.Song) error {
	if _, ok := fake.songs[song.ID]; ok {
		return general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	fake.songs[song.ID] = song
	return nil

}
func (fake testDB) AddLike(userID, songID int) error {
	if _, ok := fake.users[userID]; !ok {
		return general.GetDBError("Missing key", general.MissingForeignKey)
	}
	if _, ok := fake.songs[songID]; !ok {
		return general.GetDBError("Missing key", general.MissingForeignKey)
	}
	if _, ok := fake.likes[userID][songID]; ok {
		return general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	fake.likes[userID][songID] = fake.songs[songID]
	return nil
}

func (fake testDB) AddDislike(userID, songID int) error {
	if _, ok := fake.users[userID]; !ok {
		return general.GetDBError("Missing key", general.MissingForeignKey)
	}
	if _, ok := fake.songs[songID]; !ok {
		return general.GetDBError("Missing key", general.MissingForeignKey)
	}
	if _, ok := fake.dislikes[userID][songID]; ok {
		return general.GetDBError("Duplicate entry", general.DuplicateEntry)
	}
	fake.dislikes[userID][songID] = fake.songs[songID]
	return nil
}

func (fake testDB) RemoveLike(userID, songID int) error {
	if _, ok := fake.likes[userID][songID]; ok {
		delete(fake.likes[userID], songID)
	}
	return nil
}

func (fake testDB) RemoveDislike(userID, songID int) error {
	if _, ok := fake.dislikes[userID][songID]; ok {
		delete(fake.dislikes[userID], songID)
	}
	return nil
}

func (fake testDB) GetLikes(userID, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	likesUser := fake.likes[userID]
	songNames := make([]general.Song, 0, len(likesUser))
	for _, song := range likesUser {
		songNames = append(songNames, song)
	}
	sort.SliceStable(songNames, func(i, j int) bool {
		return songNames[i].Name < songNames[j].Name
	})
	searchResults := make([]general.Song, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(songNames)))); indexResult++ {
		likedSong := likesUser[songNames[indexResult].ID]
		likedSong.Preference = "like"
		searchResults = append(searchResults, likedSong)
	}
	return searchResults, nil
}

func (fake testDB) GetDislikes(userID, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	dislikesUser := fake.dislikes[userID]
	dislikedSongs := make([]general.Song, 0, len(dislikesUser))
	for _, song := range dislikesUser {
		dislikedSongs = append(dislikedSongs, song)
	}
	sort.SliceStable(dislikedSongs, func(i, j int) bool {
		return dislikedSongs[i].Name < dislikedSongs[j].Name
	})
	searchResults := make([]general.Song, 0, max)
	for indexResult := offset; indexResult < int(math.Min(float64(offset+max), float64(len(dislikedSongs)))); indexResult++ {
		dislikedSong := dislikesUser[dislikedSongs[indexResult].ID]
		dislikedSong.Preference = "dislike"
		searchResults = append(searchResults, dislikedSong)
	}
	return searchResults, nil
}
func (fake testDB) GetLikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup) {
	defer close(channel)
	likesUser := fake.likes[userID]
	for id, song := range likesUser {
		for _, artist := range song.Artists {
			if artist.Name == nameArtist {
				channel <- id
				continue
			}
		}
	}
	wg.Done()
}

func (fake testDB) GetDislikesIDFromArtistName(logger *log.Logger, userID int, nameArtist string, channel chan<- int, wg *sync.WaitGroup) {
	defer close(channel)
	dislikesUser := fake.dislikes[userID]
	for id, song := range dislikesUser {
		for _, artist := range song.Artists {
			if artist.Name == nameArtist {
				channel <- id
				continue
			}
		}
	}
	wg.Done()
}

type preference struct {
	user, song int
}

func newPreference(user, song int) preference {
	return preference{user: user, song: song}
}

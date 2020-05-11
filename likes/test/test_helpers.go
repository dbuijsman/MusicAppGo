package test

import (
	"fmt"
	"general"
	"io"
	"likes/database"
	"likes/handlers"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func testServer(db database.Database, existingSongs []general.Song) *http.Server {
	server, _ := handlers.NewLikesServer(testLikesHandler(db, existingSongs), nil, "likes_test", "")
	return server
}

func testLikesHandler(db database.Database, existingSongs []general.Song) *handlers.LikesHandler {
	return handlers.NewLikesHandler(general.TestEmptyLogger(), db, testGetRequest(existingSongs))
}

func testGetRequest(existingSongs []general.Song) func(string) (*http.Response, error) {
	songDB := make(map[int]general.Song)
	for _, song := range existingSongs {
		songDB[song.ID] = song
	}
	return func(address string) (*http.Response, error) {
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

func addDBToArray(missingSongs []general.Song, db testDB) []general.Song {
	for _, song := range db.songs {
		missingSongs = append(missingSongs, song)
	}
	return missingSongs
}

type preference struct {
	user, song int
}

func newPreference(user, song int) preference {
	return preference{user: user, song: song}
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

func (fake testDB) addPreferencesToTestDB(t *testing.T, userID int, songs []general.Song, prefFunction func(int, int) error) {
	fake.addSongsToTestDB(t, songs)
	for _, song := range songs {
		if err := prefFunction(userID, song.ID); err != nil {
			t.Fatalf("Failed to add preference for user #%v and song %v due to: %s\n", userID, song.Name, err)
		}
	}
}

func (fake testDB) addSongsToTestDB(t *testing.T, songs []general.Song) {
	for _, song := range songs {
		for _, artist := range song.Artists {
			fake.AddArtist(artist)
		}
		err := fake.AddSong(song)
		if err != nil {
			t.Fatalf("Failed to add song %v due to: %s\n", song.Name, err)
		}
	}
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
	for _, artist := range song.Artists {
		if _, ok := fake.artists[artist.Name]; !ok {
			return general.GetDBError("Missing foreign key", general.MissingForeignKey)
		}
	}
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
	likedSongs := make([]general.Song, 0, len(likesUser))
	for _, song := range likesUser {
		song.Preference = "like"
		likedSongs = append(likedSongs, song)
	}
	sort.SliceStable(likedSongs, func(i, j int) bool {
		return songIsSmallerThan(likedSongs[i], likedSongs[j])
	})
	return likedSongs[int(math.Min(float64(offset), float64(len(likedSongs)))):int(math.Min(float64(offset+max), float64(len(likedSongs))))], nil
}

func (fake testDB) GetDislikes(userID, offset, max int) ([]general.Song, error) {
	if max <= 0 || offset < 0 {
		return nil, general.GetDBError("Can not search with negative offset or non-positive max", general.InvalidOffsetMax)
	}
	dislikesUser := fake.dislikes[userID]
	dislikedSongs := make([]general.Song, 0, len(dislikesUser))
	for _, song := range dislikesUser {
		song.Preference = "dislike"
		dislikedSongs = append(dislikedSongs, song)
	}
	sort.SliceStable(dislikedSongs, func(i, j int) bool {
		return songIsSmallerThan(dislikedSongs[i], dislikedSongs[j])
	})
	return dislikedSongs[int(math.Min(float64(offset), float64(len(dislikedSongs)))):int(math.Min(float64(offset+max), float64(len(dislikedSongs))))], nil
}

func songIsSmallerThan(song1, song2 general.Song) bool {
	firstArtistSong1 := getFirstOrderedArtist(song1)
	firstArtistSong2 := getFirstOrderedArtist(song2)
	if firstArtistSong1.Name < firstArtistSong2.Name {
		return true
	}
	if firstArtistSong1.Name > firstArtistSong2.Name {
		return false
	}
	return song1.Name < song2.Name
}

func getFirstOrderedArtist(song general.Song) general.Artist {
	artists := make([]general.Artist, len(song.Artists))
	copy(artists, song.Artists)
	sort.SliceStable(artists, func(i, j int) bool {
		return artists[i].Name < artists[j].Name
	})
	return artists[0]
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

func seperatePrefix(name string) (artist, prefix string) {
	if len(name) < 4 {
		artist = name
		return
	}
	arrayPrefixes := []string{"A ", "An ", "The "}
	for _, entry := range arrayPrefixes {
		if name[0:len(entry)] == entry {
			prefix = strings.Trim(entry, " ")
			artist = name[len(entry):]
			return
		}
	}
	artist = name
	return
}

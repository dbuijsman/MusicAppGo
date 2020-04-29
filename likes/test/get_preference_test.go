package test

import (
	"general"
	"net/http"
	"testing"
)

func TestGetLikes_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence")
	cases := map[string]struct {
		existingLikes      []preference
		offset             int
		expectedStatusCode int
	}{
		"User has some likes":                          {[]preference{newPreference(existingUser.ID, existingSong.ID)}, 0, http.StatusOK},
		"No like is found":                             {nil, 0, http.StatusNotFound},
		"Offset is bigger than total amount of result": {[]preference{newPreference(existingUser.ID, existingSong.ID)}, 1, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(existingSong)
		for _, like := range testCase.existingLikes {
			db.AddLike(like.user, like.song)
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, existingUser)
		request = general.WithOffsetMax(request, testCase.offset, 10)
		response := general.TestSendRequest(t, handler.GetLikes, request)
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", nameCase, testCase.expectedStatusCode, response.Code)
		}
	}
}

func TestGetLikes_amountResults(t *testing.T) {
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Papa Roach", "")}, "Last Resort"),
		general.NewSong(4, []general.Artist{general.NewArtist(3, "Blink-182", "")}, "Dammit"),
	}
	cases := map[string]struct {
		users                 []general.Credentials
		likes                 []preference
		dislikes              []preference
		requestingUser        general.Credentials
		offset, max           int
		expectedAmountResults int
	}{
		"Only songs that are liked are included": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2)}, nil, general.NewCredentials(1, "Test", ""), 0, 10, 2,
		},
		"Disliked songs are excluded": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 3)},
			[]preference{newPreference(1, 1), newPreference(1, 2)}, general.NewCredentials(1, "Test", ""), 0, 10, 1,
		},
		"Only songs liked by the user": {
			[]general.Credentials{general.NewCredentials(1, "Test", ""), general.NewCredentials(2, "Other user", "")},
			[]preference{newPreference(1, 1), newPreference(2, 2), newPreference(2, 1)}, nil, general.NewCredentials(1, "Test", ""), 0, 10, 1,
		},
		"Amount of results capped by max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3)}, nil, general.NewCredentials(1, "Test", ""), 0, 2, 2,
		},
		"Skip results given by max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3)}, nil, general.NewCredentials(1, "Test", ""), 1, 10, 2,
		},
		"Correctly handle combination of offset and max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3), newPreference(1, 4)}, nil,
			general.NewCredentials(1, "Test", ""), 1, 2, 2,
		},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		for _, user := range testCase.users {
			db.AddUser(user)
		}
		for _, song := range songs {
			db.AddSong(song)
		}
		for _, like := range testCase.likes {
			db.AddLike(like.user, like.song)
		}
		for _, dislike := range testCase.dislikes {
			db.AddDislike(dislike.user, dislike.song)
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, testCase.requestingUser)
		request = general.WithOffsetMax(request, testCase.offset, testCase.max)
		response := general.TestSendRequest(t, handler.GetLikes, request)
		var result general.MultipleSongs
		err = general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
			continue
		}
		if len(result.Data) != testCase.expectedAmountResults {
			t.Errorf("%v: Searching for songs for user %v, offset %v and max %v should give %v results but got %v\n", nameCase, testCase.requestingUser.Username, testCase.offset, testCase.max, testCase.expectedAmountResults, len(result.Data))
		}
	}
}

func TestGetLikes_HasNextResult(t *testing.T) {
	user := general.NewCredentials(1, "Test", "")
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Papa Roach", "")}, "Last Resort"),
		general.NewSong(4, []general.Artist{general.NewArtist(3, "Blink-182", "")}, "Dammit"),
		general.NewSong(5, []general.Artist{general.NewArtist(4, "Led Zeppelin", "")}, "Stairway to Heaven"),
	}
	cases := map[string]struct {
		likedSongs      []int
		offset, max     int
		expectedHasNext bool
	}{
		"Max bigger than amount of results":           {[]int{1, 2}, 0, 10, false},
		"Offset + max smaller than amount of results": {[]int{1, 2, 3, 4}, 0, 2, true},
		"Offset + max bigger than amount of results":  {[]int{1, 2, 3, 4}, 3, 3, false},
		"Offset + max equal to amount of results":     {[]int{1, 2, 3, 4}, 2, 2, false},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(user)
		for _, song := range songs {
			db.AddSong(song)
		}
		for _, songID := range testCase.likedSongs {
			db.AddLike(user.ID, songID)
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, user)
		request = general.WithOffsetMax(request, testCase.offset, testCase.max)
		response := general.TestSendRequest(t, handler.GetLikes, request)
		var result general.MultipleSongs
		err = general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
			continue
		}
		if hasNext := result.HasNext; hasNext != testCase.expectedHasNext {
			t.Errorf("%v: Expects hasNext equal to: %v but got: %v\n", nameCase, testCase.expectedHasNext, hasNext)
		}
	}
}

func TestGetLikes_withLikeTag(t *testing.T) {
	user := general.NewCredentials(1, "Test", "")
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Last Frequencies", ""), general.NewArtist(3, "Zonderling", "")}, "Crazy"),
		general.NewSong(4, []general.Artist{general.NewArtist(4, "Blink-182", "")}, "Dammit"),
		general.NewSong(5, []general.Artist{general.NewArtist(5, "Led Zeppelin", "")}, "Stairway to Heaven"),
	}
	db := newTestDB()
	db.AddUser(user)
	for _, song := range songs {
		db.AddSong(song)
		db.AddLike(user.ID, song.ID)
	}
	handler := testLikesHandlerNilRequest(db)
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("With like tag: Failed to create request due to: %v\n", err)
	}
	request = general.WithCredentials(request, user)
	request = general.WithOffsetMax(request, 0, 10)
	response := general.TestSendRequest(t, handler.GetLikes, request)
	var result general.MultipleSongs
	err = general.ReadFromJSON(&result, response.Body)
	if err != nil {
		t.Fatalf("[ERROR] With like tag: Decoding response: %v\n", err)
	}
	for _, song := range result.Data {
		if song.Preference != "like" {
			t.Errorf("With like tag: %v -%v: Expects preference equals like but got: %v\n", song.Artists, song.Name, song.Preference)
		}
	}
}

func TestGetDislikes_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence")
	cases := map[string]struct {
		existingDislikes   []preference
		offset             int
		expectedStatusCode int
	}{
		"User has some dislikes":                       {[]preference{newPreference(existingUser.ID, existingSong.ID)}, 0, http.StatusOK},
		"No dislike is found":                          {nil, 0, http.StatusNotFound},
		"Offset is bigger than total amount of result": {[]preference{newPreference(existingUser.ID, existingSong.ID)}, 1, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(existingSong)
		for _, dislike := range testCase.existingDislikes {
			err := db.AddDislike(dislike.user, dislike.song)
			if err != nil {
				t.Errorf("[ERROR] %v: Can't add dislike due to: %v\n", nameCase, err)
			}
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, existingUser)
		request = general.WithOffsetMax(request, testCase.offset, 10)
		response := general.TestSendRequest(t, handler.GetDislikes, request)
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", nameCase, testCase.expectedStatusCode, response.Code)
		}
	}
}

func TestGetDislikes_amountResults(t *testing.T) {
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Papa Roach", "")}, "Last Resort"),
		general.NewSong(4, []general.Artist{general.NewArtist(3, "Blink-182", "")}, "Dammit"),
	}
	cases := map[string]struct {
		users                 []general.Credentials
		dislikes              []preference
		likes                 []preference
		requestingUser        general.Credentials
		offset, max           int
		expectedAmountResults int
	}{
		"Only songs that are disliked are included": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2)}, nil, general.NewCredentials(1, "Test", ""), 0, 10, 2,
		},
		"Liked songs are excluded": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 3)},
			[]preference{newPreference(1, 1), newPreference(1, 2)}, general.NewCredentials(1, "Test", ""), 0, 10, 1,
		},
		"Only songs disliked by the user": {
			[]general.Credentials{general.NewCredentials(1, "Test", ""), general.NewCredentials(2, "Other user", "")},
			[]preference{newPreference(1, 1), newPreference(2, 2), newPreference(2, 1)}, nil, general.NewCredentials(1, "Test", ""), 0, 10, 1,
		},
		"Amount of results capped by max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3)}, nil, general.NewCredentials(1, "Test", ""), 0, 2, 2,
		},
		"Skip results given by max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3)}, nil, general.NewCredentials(1, "Test", ""), 1, 10, 2,
		},
		"Correctly handle combination of offset and max": {
			[]general.Credentials{general.NewCredentials(1, "Test", "")},
			[]preference{newPreference(1, 1), newPreference(1, 2), newPreference(1, 3), newPreference(1, 4)}, nil,
			general.NewCredentials(1, "Test", ""), 1, 2, 2,
		},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		for _, user := range testCase.users {
			db.AddUser(user)
		}
		for _, song := range songs {
			db.AddSong(song)
		}
		for _, dislike := range testCase.dislikes {
			db.AddDislike(dislike.user, dislike.song)
		}
		for _, like := range testCase.likes {
			db.AddLike(like.user, like.song)
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, testCase.requestingUser)
		request = general.WithOffsetMax(request, testCase.offset, testCase.max)
		response := general.TestSendRequest(t, handler.GetDislikes, request)
		var result general.MultipleSongs
		err = general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
			continue
		}
		if len(result.Data) != testCase.expectedAmountResults {
			t.Errorf("%v: Searching for songs for user %v, offset %v and max %v should give %v results but got %v\n", nameCase, testCase.requestingUser.Username, testCase.offset, testCase.max, testCase.expectedAmountResults, len(result.Data))
		}
	}
}

func TestGetDislikes_HasNextResult(t *testing.T) {
	user := general.NewCredentials(1, "Test", "")
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Papa Roach", "")}, "Last Resort"),
		general.NewSong(4, []general.Artist{general.NewArtist(3, "Blink-182", "")}, "Dammit"),
		general.NewSong(5, []general.Artist{general.NewArtist(4, "Led Zeppelin", "")}, "Stairway to Heaven"),
	}
	cases := map[string]struct {
		dislikedSongs   []int
		offset, max     int
		expectedHasNext bool
	}{
		"Max bigger than amount of results":           {[]int{1, 2}, 0, 10, false},
		"Offset + max smaller than amount of results": {[]int{1, 2, 3, 4}, 0, 2, true},
		"Offset + max bigger than amount of results":  {[]int{1, 2, 3, 4}, 3, 3, false},
		"Offset + max equal to amount of results":     {[]int{1, 2, 3, 4}, 2, 2, false},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(user)
		for _, song := range songs {
			db.AddSong(song)
		}
		for _, songID := range testCase.dislikedSongs {
			db.AddDislike(user.ID, songID)
		}
		handler := testLikesHandlerNilRequest(db)
		request, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Errorf("%v: Failed to create request due to: %v\n", nameCase, err)
			continue
		}
		request = general.WithCredentials(request, user)
		request = general.WithOffsetMax(request, testCase.offset, testCase.max)
		response := general.TestSendRequest(t, handler.GetDislikes, request)
		var result general.MultipleSongs
		err = general.ReadFromJSON(&result, response.Body)
		if err != nil {
			t.Errorf("[ERROR] %v: Decoding response: %v\n", nameCase, err)
			continue
		}
		if hasNext := result.HasNext; hasNext != testCase.expectedHasNext {
			t.Errorf("%v: Expects hasNext equal to: %v but got: %v\n", nameCase, testCase.expectedHasNext, hasNext)
		}
	}
}

func TestGetDislikes_withDislikeTag(t *testing.T) {
	user := general.NewCredentials(1, "Test", "")
	songs := []general.Song{
		general.NewSong(1, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "The Sound of Silence"),
		general.NewSong(2, []general.Artist{general.NewArtist(1, "Disturbed", "")}, "Land of Confusion"),
		general.NewSong(3, []general.Artist{general.NewArtist(2, "Last Frequencies", ""), general.NewArtist(3, "Zonderling", "")}, "Crazy"),
		general.NewSong(4, []general.Artist{general.NewArtist(4, "Blink-182", "")}, "Dammit"),
		general.NewSong(5, []general.Artist{general.NewArtist(5, "Led Zeppelin", "")}, "Stairway to Heaven"),
	}
	db := newTestDB()
	db.AddUser(user)
	for _, song := range songs {
		db.AddSong(song)
		db.AddDislike(user.ID, song.ID)
	}
	handler := testLikesHandlerNilRequest(db)
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("With dislike tag: Failed to create request due to: %v\n", err)
	}
	request = general.WithCredentials(request, user)
	request = general.WithOffsetMax(request, 0, 10)
	response := general.TestSendRequest(t, handler.GetDislikes, request)
	var result general.MultipleSongs
	err = general.ReadFromJSON(&result, response.Body)
	if err != nil {
		t.Fatalf("[ERROR] With dislike tag: Decoding response: %v\n", err)
	}
	for _, song := range result.Data {
		if song.Preference != "dislike" {
			t.Errorf("With dislike tag: %v -%v: Expects preference equals dislike but got: %v\n", song.Artists, song.Name, song.Preference)
		}
	}
}

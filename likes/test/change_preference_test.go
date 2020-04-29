package test

import (
	"general"
	"net/http"
	"testing"
	"time"
)

func TestAddLike_saveInDB(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSongs := []general.Song{general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep"),
		general.NewSong(2, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Duality")}
	missedSong := general.NewSong(3, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Snuff")
	cases := map[string]struct {
		userID, songID    int
		expectedSavedInDB bool
	}{
		"Complete existing data": {existingUser.ID, existingSongs[0].ID, true},
		"New user":               {404, existingSongs[0].ID, true},
		"Missed song":            {existingUser.ID, missedSong.ID, true},
		"Not existing song":      {existingUser.ID, 404, false},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.addSongsToTestDB(existingSongs)
		handler := testLikesHandlerWithRequest(db, append(existingSongs, missedSong))
		general.TestPostRequestWithContext(t, handler.AddLike, general.Preference{ID: testCase.songID}, general.Credentials{}, general.NewCredentials(testCase.userID, "Test", ""))
		if _, ok := db.likes[testCase.userID][testCase.songID]; ok != testCase.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved in db: %v but got %v\n", nameCase, testCase.expectedSavedInDB, ok)
		}
	}
}

func TestAddLike_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSongs := []general.Song{general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep"),
		general.NewSong(2, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Duality")}
	missedSong := general.NewSong(3, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Snuff")
	cases := map[string]struct {
		userID, songID     int
		expectedStatusCode int
	}{
		"Complete existing data": {existingUser.ID, existingSongs[0].ID, http.StatusOK},
		"New user":               {404, existingSongs[0].ID, http.StatusOK},
		"Missed song":            {existingUser.ID, missedSong.ID, http.StatusOK},
		"Not existing song":      {existingUser.ID, 404, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.addSongsToTestDB(existingSongs)
		handler := testLikesHandlerWithRequest(db, append(existingSongs, missedSong))
		response := general.TestPostRequestWithContext(t, handler.AddLike, general.Preference{ID: testCase.songID}, general.Credentials{}, general.NewCredentials(testCase.userID, "Test", ""))
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got %v\n", nameCase, testCase.expectedStatusCode, response.Code)
		}
	}
}

func TestAddLike_removeDislike(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	db := newTestDB()
	db.AddUser(existingUser)
	db.AddSong(existingSong)
	err := db.AddDislike(existingUser.ID, existingSong.ID)
	if err != nil {
		t.Fatalf("Can't add dislike due to: %s\n", err)
	}
	handler := testLikesHandlerNilRequest(db)
	general.TestPostRequestWithContext(t, handler.AddLike, general.Preference{ID: existingSong.ID}, general.Credentials{}, existingUser)
	time.Sleep(time.Millisecond)
	if _, ok := db.dislikes[existingUser.ID][existingSong.ID]; ok {
		t.Errorf("Adding a like after disliking the same song should remove the dislike\n")
	}
}

func TestAddDislike_saveInDB(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSongs := []general.Song{general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep"),
		general.NewSong(2, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Duality")}
	missedSong := general.NewSong(3, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Snuff")
	cases := map[string]struct {
		userID, songID    int
		expectedSavedInDB bool
	}{
		"Complete existing data": {existingUser.ID, existingSongs[0].ID, true},
		"New user":               {404, existingSongs[0].ID, true},
		"Missed song":            {existingUser.ID, missedSong.ID, true},
		"Not existing song":      {existingUser.ID, 404, false},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.addSongsToTestDB(existingSongs)
		handler := testLikesHandlerWithRequest(db, append(existingSongs, missedSong))
		general.TestPostRequestWithContext(t, handler.AddDislike, general.Preference{ID: testCase.songID}, general.Credentials{}, general.NewCredentials(testCase.userID, "Test", ""))
		if _, ok := db.dislikes[testCase.userID][testCase.songID]; ok != testCase.expectedSavedInDB {
			t.Errorf("%v: Expects to be saved in db: %v but got %v\n", nameCase, testCase.expectedSavedInDB, ok)
		}
	}
}

func TestAddDislike_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSongs := []general.Song{general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep"),
		general.NewSong(2, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Duality")}
	missedSong := general.NewSong(3, []general.Artist{general.NewArtist(2, "Slipknot", "")}, "Snuff")
	cases := map[string]struct {
		userID, songID     int
		expectedStatusCode int
	}{
		"Complete existing data": {existingUser.ID, existingSongs[0].ID, http.StatusOK},
		"New user":               {404, existingSongs[0].ID, http.StatusOK},
		"Missed song":            {existingUser.ID, missedSong.ID, http.StatusOK},
		"Not existing song":      {existingUser.ID, 404, http.StatusNotFound},
	}
	for nameCase, testCase := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.addSongsToTestDB(existingSongs)
		handler := testLikesHandlerWithRequest(db, append(existingSongs, missedSong))
		response := general.TestPostRequestWithContext(t, handler.AddDislike, general.Preference{ID: testCase.songID}, general.Credentials{}, general.NewCredentials(testCase.userID, "Test", ""))
		if response.Code != testCase.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got %v\n", nameCase, testCase.expectedStatusCode, response.Code)
		}
	}
}

func TestAddDislike_removeLike(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	existingSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	db := newTestDB()
	db.AddUser(existingUser)
	db.AddSong(existingSong)
	err := db.AddLike(existingUser.ID, existingSong.ID)
	if err != nil {
		t.Fatalf("Can't add like due to: %s\n", err)
	}
	handler := testLikesHandlerNilRequest(db)
	general.TestPostRequestWithContext(t, handler.AddDislike, general.Preference{ID: existingSong.ID}, general.Credentials{}, existingUser)
	time.Sleep(time.Millisecond)
	if _, ok := db.likes[existingUser.ID][existingSong.ID]; ok {
		t.Errorf("Adding a dislike after liking the same song should remove the like\n")
	}
}

func TestRemoveLike_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	likedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	cases := map[string]struct {
		songID             int
		expectedStatusCode int
	}{
		"Existing like":     {likedSong.ID, http.StatusOK},
		"Non-existing like": {404, http.StatusOK},
	}
	for nameCase, caseLike := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(likedSong)
		err := db.AddLike(existingUser.ID, likedSong.ID)
		if err != nil {
			t.Fatalf("Can't add like due to: %s\n", err)
		}
		handler := testLikesHandlerNilRequest(db)
		response := general.TestPostRequestWithContext(t, handler.RemoveLike, general.Preference{ID: caseLike.songID}, general.Credentials{}, existingUser)
		if response.Code != caseLike.expectedStatusCode {
			t.Errorf("%v: Expects statuscode %v but got: %v\n", nameCase, response.Code, caseLike.expectedStatusCode)
		}
	}
}

func TestRemoveLike_removeFromDB(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	likedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	cases := map[string]struct {
		songID int
	}{
		"Existing like":     {likedSong.ID},
		"Non-existing like": {404},
	}
	for nameCase, caseLike := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(likedSong)
		err := db.AddLike(existingUser.ID, likedSong.ID)
		if err != nil {
			t.Fatalf("Can't add like due to: %s\n", err)
		}
		handler := testLikesHandlerNilRequest(db)
		general.TestPostRequestWithContext(t, handler.RemoveLike, general.Preference{ID: caseLike.songID}, general.Credentials{}, existingUser)
		if _, ok := db.likes[existingUser.ID][caseLike.songID]; ok {
			t.Errorf("%v: Expects that a like was removed but this didn't happen\n", nameCase)
		}
	}
}

func TestRemoveLike_dislikedSong(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	dislikedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	db := newTestDB()
	db.AddUser(existingUser)
	db.AddSong(dislikedSong)
	err := db.AddDislike(existingUser.ID, dislikedSong.ID)
	if err != nil {
		t.Fatalf("Can't add dislike due to: %s\n", err)
	}
	handler := testLikesHandlerNilRequest(db)
	general.TestPostRequestWithContext(t, handler.RemoveLike, general.Preference{ID: dislikedSong.ID}, general.Credentials{}, existingUser)
	if _, ok := db.dislikes[existingUser.ID][dislikedSong.ID]; !ok {
		t.Errorf("Remove a like while the song was disliked expects taht nothing happens but the dislike was removed\n")
	}
}

func TestRemoveDislike_statusCode(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	dislikedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	cases := map[string]struct {
		songID             int
		expectedStatusCode int
	}{
		"Existing dislike":     {dislikedSong.ID, http.StatusOK},
		"Non-existing dislike": {404, http.StatusOK},
	}
	for nameCase, caseDislike := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(dislikedSong)
		err := db.AddDislike(existingUser.ID, dislikedSong.ID)
		if err != nil {
			t.Fatalf("Can't add dislike due to: %s\n", err)
		}
		handler := testLikesHandlerNilRequest(db)
		response := general.TestPostRequestWithContext(t, handler.RemoveDislike, general.Preference{ID: caseDislike.songID}, general.Credentials{}, existingUser)
		if response.Code != caseDislike.expectedStatusCode {
			t.Errorf("%v: Expects statuscode %v but got: %v\n", nameCase, response.Code, caseDislike.expectedStatusCode)
		}
	}
}

func TestRemoveDislike_removeFromDB(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	dislikedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	cases := map[string]struct {
		songID int
	}{
		"Existing dislike":     {dislikedSong.ID},
		"Non-existing dislike": {404},
	}
	for nameCase, caseDislike := range cases {
		db := newTestDB()
		db.AddUser(existingUser)
		db.AddSong(dislikedSong)
		err := db.AddDislike(existingUser.ID, dislikedSong.ID)
		if err != nil {
			t.Fatalf("Can't add dislike due to: %s\n", err)
		}
		handler := testLikesHandlerNilRequest(db)
		general.TestPostRequestWithContext(t, handler.RemoveDislike, general.Preference{ID: caseDislike.songID}, general.Credentials{}, existingUser)
		if _, ok := db.likes[existingUser.ID][caseDislike.songID]; ok {
			t.Errorf("%v: Expects that a like was removed but this didn't happen\n", nameCase)
		}
	}
}

func TestRemoveDislike_likedSong(t *testing.T) {
	existingUser := general.NewCredentials(1, "Test", "")
	likedSong := general.NewSong(1, []general.Artist{general.NewArtist(1, "Sum 41", "")}, "In Too Deep")
	db := newTestDB()
	db.AddUser(existingUser)
	db.AddSong(likedSong)
	err := db.AddLike(existingUser.ID, likedSong.ID)
	if err != nil {
		t.Fatalf("Can't add like due to: %s\n", err)
	}
	handler := testLikesHandlerNilRequest(db)
	general.TestPostRequestWithContext(t, handler.RemoveDislike, general.Preference{ID: likedSong.ID}, general.Credentials{}, existingUser)
	if _, ok := db.likes[existingUser.ID][likedSong.ID]; !ok {
		t.Errorf("Remove a dislike while the song was liked expects taht nothing happens but the like was removed\n")
	}
}

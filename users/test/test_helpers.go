package test

import (
	"MusicAppGo/common"
	"log"
	"net/http"
	"net/http/httptest"
	"users/database"
	"users/handlers"
)

func testUserHandler() *handlers.UserHandler {
	l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
	return handlers.NewUserHandler(l, testDB{db: make(map[string]string)}, func(topic string, message []byte) error {
		return nil
	})
}

type testWriter struct{}

func (fake testWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	return
}

type testDB struct {
	db map[string]string
}

func (fake testDB) InsertUser(username, password string) error {
	if _, ok := fake.db[username]; ok {
		err := common.GetDBError("Dublicate entry", common.DuplicateEntry)
		return err
	}
	fake.db[username] = password
	return nil
}
func (fake testDB) Login(username, password string) (bool, error) {
	entry, ok := fake.db[username]
	if !ok {
		return false, nil
	}
	return entry == password, nil
}
func (fake testDB) FindUser(username string) (database.RowUserDB, error) {
	password, ok := fake.db[username]
	if !ok {
		return *database.NewRowUserDB("", ""), nil
	}
	return *database.NewRowUserDB(username, password), nil
}

type testHandler struct{}

func (handler testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func testValidateToken(token string) bool {
	handler := testHandler{}
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Add("Token", token)
	recorder := httptest.NewRecorder()
	tokenValidator := common.GetValidateTokenMiddleWare(log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile))
	tokenValidator(handler).ServeHTTP(recorder, request)
	return recorder.Code == http.StatusOK
}

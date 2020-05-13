package test

import (
	"general/dberror"
	"general/server"
	"general/testhelpers"
	"general/types"
	"net/http"
	"net/http/httptest"
	"testing"
	"user_data/database"
	"user_data/handlers"
)

func testServer(t *testing.T, db database.Database) (*http.Server, chan testhelpers.Message) {
	sendMessage, channel := testhelpers.TestSendMessage()
	handler, err := handlers.NewUserHandler(testhelpers.TestEmptyLogger(), db, sendMessage)
	if err != nil {
		t.Fatalf("Failed to create a testServer due to: %s\n", err)
	}
	newServer, _ := handlers.NewUserServer(handler, nil, "user_data_test", "", 0)
	return newServer, channel
}

type testCredentials struct {
	id                 int
	username, password string
}

type testDB struct {
	db map[string]testCredentials
}

func newTestDB() testDB {
	return testDB{db: make(map[string]testCredentials)}
}

func (fake testDB) SignUp(username, password string) (int, error) {
	if _, ok := fake.db[username]; ok {
		err := dberror.GetDBError("Dublicate entry", dberror.DuplicateEntry)
		return 0, err
	}
	id := len(fake.db) + 1
	fake.db[username] = testCredentials{id: id, username: username, password: password}
	return id, nil
}
func (fake testDB) Login(username, password string) (types.Credentials, error) {
	entry, ok := fake.db[username]
	if !ok || entry.password != password {
		return types.Credentials{}, dberror.GetDBError("The credentials do not match", dberror.InvalidInput)
	}
	return types.NewCredentials(entry.id, entry.username, ""), nil
}
func (fake testDB) FindUser(username string) (types.Credentials, error) {
	user, ok := fake.db[username]
	if !ok {
		return types.Credentials{}, dberror.GetDBError("Can't find user", dberror.NotFoundError)
	}
	return types.NewCredentials(user.id, user.username, ""), nil
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
	tokenValidator := server.GetValidateTokenMiddleWare(testhelpers.TestEmptyLogger())
	tokenValidator(handler).ServeHTTP(recorder, request)
	return recorder.Code == http.StatusOK
}

package test

import (
	"general"
	"net/http"
	"net/http/httptest"
	"user_data/handlers"
)

func testUserHandler() *handlers.UserHandler {
	l := general.TestEmptyLogger()
	return handlers.NewUserHandler(l, testDB{db: make(map[string]testCredentials)}, general.TestSendMessageEmpty())
}

type testDB struct {
	db map[string]testCredentials
}
type testCredentials struct {
	id                 int
	username, password string
}

func (fake testDB) SignUp(username, password string) (int, error) {
	if _, ok := fake.db[username]; ok {
		err := general.GetDBError("Dublicate entry", general.DuplicateEntry)
		return 0, err
	}
	id := len(fake.db) + 1
	fake.db[username] = testCredentials{id: id, username: username, password: password}
	return id, nil
}
func (fake testDB) Login(username, password string) (general.Credentials, error) {
	entry, ok := fake.db[username]
	if !ok || entry.password != password {
		return general.Credentials{}, general.GetDBError("The credentials do not match", general.InvalidInput)
	}
	return general.NewCredentials(entry.id, entry.username, ""), nil
}
func (fake testDB) FindUser(username string) (general.Credentials, error) {
	user, ok := fake.db[username]
	if !ok {
		return general.Credentials{}, general.GetDBError("Can't find user", general.NotFoundError)
	}
	return general.NewCredentials(user.id, user.username, ""), nil
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
	tokenValidator := general.GetValidateTokenMiddleWare(general.TestEmptyLogger())
	tokenValidator(handler).ServeHTTP(recorder, request)
	return recorder.Code == http.StatusOK
}

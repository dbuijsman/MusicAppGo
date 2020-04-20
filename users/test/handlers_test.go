package test

import (
	"MusicAppGo/common"
	"log"
	"net/http"
	"sync"
	"testing"
	"users/handlers"
)

func TestSignUp_statusCode(t *testing.T) {
	cases := map[string]struct {
		username, password string
		expected           int
	}{
		"Complete credentials": {"Test", "Password", http.StatusOK},
		"Missing username":     {"", "Password", http.StatusBadRequest},
		"Missing password":     {"Test", "", http.StatusBadRequest},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler().SignUp
		recorder := common.TestPostRequest(t, handler, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		if recorder.Code != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v expects statuscode: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, recorder.Code)
		}
	}
}
func TestSignUp_savingInDB(t *testing.T) {
	cases := map[string]struct {
		username, password string
		expected           bool
	}{
		"Complete credentials": {"Test", "Password", true},
		"Missing username":     {"", "Password", false},
		"Missing password":     {"Test", "", false},
	}
	for nameCase, credentials := range cases {
		l := log.New(testWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
		db := testDB{db: make(map[string]string)}
		handler := handlers.NewUserHandler(l, db, func(topic string, message []byte) error {
			return nil
		}).SignUp
		common.TestPostRequest(t, handler, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		_, result := db.db[credentials.username]
		if result != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v expects saving in DB: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, result)
		}
	}
}
func TestSignUp_returningValidToken(t *testing.T) {
	cases := map[string]struct {
		username, password string
		expected           bool
	}{
		"Complete credentials": {"Test", "Password", true},
		"Missing username":     {"", "Password", false},
		"Missing password":     {"Test", "", false},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler().SignUp
		recorder := common.TestPostRequest(t, handler, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		result := testValidateToken(recorder.Body.String())
		if result != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v expects valid token: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, result)
		}
	}
}
func TestSignUp_duplicateEntry(t *testing.T) {
	creds := handlers.Credentials{Username: "Test", Password: "Testing"}
	cases := map[string]struct {
		username, password string
		expected           int
	}{
		"Duplicate credentials": {creds.Username, creds.Password, http.StatusUnprocessableEntity},
		"Duplicate password":    {creds.Username + "NOT", "Password", http.StatusOK},
		"Duplicate username":    {creds.Username, creds.Password + "NOT", http.StatusUnprocessableEntity},
		"Different credentials": {creds.Username + "NOT", creds.Password + "NOT", http.StatusOK},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler().SignUp
		common.TestPostRequest(t, handler, creds)
		recorder := common.TestPostRequest(t, handler, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		if recorder.Code != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v after base case expects statuscode: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, recorder.Code)
		}
	}
}
func TestSignUp_sendMessage(t *testing.T) {
	topicValue, msgValue := "", ""
	top, msg := &topicValue, &msgValue
	creds := handlers.Credentials{Username: "Test", Password: "Testing"}
	expectedTopic, expectedMessage := "signup", creds.Username
	handler := testUserHandler()
	var wg sync.WaitGroup
	wg.Add(1)
	handler.SendMessage = func(topic string, message []byte) {
		*top = topic
		*msg = string(message)
		wg.Done()
	}
	common.TestPostRequest(t, handler.SignUp, creds)
	wg.Wait()
	if *top != expectedTopic {
		t.Errorf("Signup expects to send a message to topic %v but instead it was send to %v\n", expectedTopic, *top)
	}
	if *msg != expectedMessage {
		t.Errorf("Signup expects to send username %v as message but instead it sends: %v\n", expectedMessage, *msg)
	}
}
func TestLogin_statusCode(t *testing.T) {
	creds := handlers.Credentials{Username: "Test", Password: "Testing"}
	cases := map[string]struct {
		username, password string
		expected           int
	}{
		"Correct credentials": {creds.Username, creds.Password, http.StatusOK},
		"Wrong username":      {creds.Username + "NOT", creds.Password, http.StatusUnauthorized},
		"Wrong password":      {creds.Username, creds.Password + "NOT", http.StatusUnauthorized},
		"Missing username":    {"", creds.Password, http.StatusBadRequest},
		"Missing password":    {creds.Username, "", http.StatusBadRequest},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler()
		common.TestPostRequest(t, handler.SignUp, creds)
		recorder := common.TestPostRequest(t, handler.Login, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		if recorder.Code != credentials.expected {
			t.Errorf("%v: Login with username: %v and password: %v after base case expects statuscode: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, recorder.Code)
		}
	}
}
func TestLogin_returningValidToken(t *testing.T) {
	creds := handlers.Credentials{Username: "Test", Password: "Testing"}
	cases := map[string]struct {
		username, password string
		expected           bool
	}{
		"Correct credentials": {creds.Username, creds.Password, true},
		"Wrong username":      {creds.Username + "NOT", creds.Password, false},
		"Wrong password":      {creds.Username, creds.Password + "NOT", false},
		"Missing username":    {"", creds.Password, false},
		"Missing password":    {creds.Username, "", false},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler()
		common.TestPostRequest(t, handler.SignUp, creds)
		recorder := common.TestPostRequest(t, handler.Login, handlers.Credentials{Username: credentials.username, Password: credentials.password})
		result := testValidateToken(recorder.Body.String())
		if result != credentials.expected {
			t.Errorf("%v, Login with username: %v and password: %v after base case expects valid token: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, result)
		}
	}
}
func TestLogin_sendMessage(t *testing.T) {
	topicValue, msgValue := "", ""
	top, msg := &topicValue, &msgValue
	creds := handlers.Credentials{Username: "Test", Password: "Testing"}
	expectedTopic, expectedMessage := "login", creds.Username
	handler := testUserHandler()
	common.TestPostRequest(t, handler.SignUp, creds)
	var wg sync.WaitGroup
	wg.Add(1)
	handler.SendMessage = func(topic string, message []byte) {
		*top = topic
		*msg = string(message)
		wg.Done()
	}
	common.TestPostRequest(t, handler.Login, creds)
	wg.Wait()
	if *top != expectedTopic {
		t.Errorf("Login expects to send a message to topic %v but instead it was send to %v\n", expectedTopic, *top)
	}
	if *msg != expectedMessage {
		t.Errorf("Login expects to send username %v as message but instead it sends: %v\n", expectedMessage, *msg)
	}
}

package test

import (
	"general"
	"net/http"
	"sync"
	"testing"
	"user_data/handlers"
)

func TestSignUp_statusCode(t *testing.T) {
	cases := map[string]struct {
		username, password string
		expected           int
	}{
		"Complete credentials": {"TestComplete", "Password", http.StatusOK},
		"Missing username":     {"", "Password", http.StatusBadRequest},
		"Missing password":     {"Test", "", http.StatusBadRequest},
	}
	for nameCase, credentials := range cases {
		handler := testUserHandler()
		recorder := general.TestPostRequest(t, handler.SignUp, handlers.NewClientCredentials(credentials.username, credentials.password))
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
		"Complete credentials": {"TestSave", "Password", true},
		"Missing username":     {"", "Password", false},
		"Missing password":     {"Test", "", false},
	}
	for nameCase, credentials := range cases {
		l := general.TestEmptyLogger()
		db := testDB{db: make(map[string]testCredentials)}
		sendMessage := general.TestSendMessageEmpty()
		handler := handlers.NewUserHandler(l, db, sendMessage).SignUp
		general.TestPostRequest(t, handler, handlers.NewClientCredentials(credentials.username, credentials.password))
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
		recorder := general.TestPostRequest(t, handler, handlers.NewClientCredentials(credentials.username, credentials.password))
		result := testValidateToken(recorder.Body.String())
		if result != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v expects valid token: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, result)
		}
	}
}

func TestSignUp_duplicateEntry(t *testing.T) {
	creds := handlers.NewClientCredentials("UserTest", "PassTest")
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
		general.TestPostRequest(t, handler, creds)
		recorder := general.TestPostRequest(t, handler, handlers.NewClientCredentials(credentials.username, credentials.password))
		if recorder.Code != credentials.expected {
			t.Errorf("%v: Signup with username: %v and password: %v after base case expects statuscode: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, recorder.Code)
		}
	}
}

func TestSignUp_sendMessage(t *testing.T) {
	creds := handlers.NewClientCredentials("UserTest", "PassTest")
	expectedTopic, expectedMessage := "newUser", creds.Username
	handler := testUserHandler()
	var wg sync.WaitGroup
	top, msg, sendMessage := general.TestSendMessage(&wg)
	handler.SendMessage = sendMessage
	wg.Add(1)
	general.TestPostRequest(t, handler.SignUp, creds)
	wg.Wait()
	var result general.Credentials
	err := general.FromJSONBytes(&result, []byte(*msg))
	if err != nil {
		t.Errorf("SignUp_sendMessage: Expects to send a message containing an user but deserializing results in: %v\n", err)
	}
	if *top != expectedTopic {
		t.Errorf("Signup expects to send a message to topic %v but instead it was send to %v\n", expectedTopic, *top)
	}
	if result.Username != expectedMessage {
		t.Errorf("Signup expects to send username %v as message but instead it sends: %v\n", expectedMessage, *msg)
	}
}

func TestLogin_statusCode(t *testing.T) {
	creds := handlers.NewClientCredentials("UserTest", "PassTest")
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
		general.TestPostRequest(t, handler.SignUp, creds)
		recorder := general.TestPostRequest(t, handler.Login, handlers.NewClientCredentials(credentials.username, credentials.password))
		if recorder.Code != credentials.expected {
			t.Errorf("%v: Login with username: %v and password: %v after base case expects statuscode: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, recorder.Code)
		}
	}
}

func TestLogin_returningValidToken(t *testing.T) {
	creds := handlers.NewClientCredentials("UserTest", "PassTest")
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
		general.TestPostRequest(t, handler.SignUp, creds)
		recorder := general.TestPostRequest(t, handler.Login, handlers.NewClientCredentials(credentials.username, credentials.password))
		result := testValidateToken(recorder.Body.String())
		if result != credentials.expected {
			t.Errorf("%v, Login with username: %v and password: %v after base case expects valid token: %v but got: %v\n", nameCase, credentials.username, credentials.password, credentials.expected, result)
		}
	}
}

func TestLogin_sendMessage(t *testing.T) {
	creds := handlers.NewClientCredentials("UserTest", "PassTest")
	topic, expectedMessage := "login", creds.Username
	handler := testUserHandler()
	general.TestPostRequest(t, handler.SignUp, creds)
	var wg sync.WaitGroup
	msg, sendMessage := general.TestSendMessageToParticularTopic(&wg, topic)
	handler.SendMessage = sendMessage
	wg.Add(1)
	general.TestPostRequest(t, handler.Login, creds)
	wg.Wait()
	var result general.Credentials
	err := general.FromJSONBytes(&result, []byte(*msg))
	if err != nil {
		t.Errorf("Login_sendMessage: Expects to send a message containing an user but deserializing results in: %v\n", err)
	}
	if result.Username != expectedMessage {
		t.Errorf("Login expects to send username %v as message but instead it sends: %v\n", expectedMessage, *msg)
	}
}

package test

import (
	"general/convert"
	"general/testhelpers"
	"general/types"
	"net/http"
	"testing"
	"time"
	"user_data/handlers"
)

func TestHandlers_response(t *testing.T) {
	user := handlers.NewClientCredentials("ExistTest", "Passexist")
	cases := map[string]struct {
		path, username, password string
		expectedStatusCode       int
		expectedValidToken       bool
	}{
		"Signup: Complete credentials":                       {"/signup", "TestComplete", "Password", http.StatusOK, true},
		"Signup: Missing username":                           {"/signup", "", "Password", http.StatusBadRequest, false},
		"Signup: Missing password":                           {"/signup", "TestComplete", "", http.StatusBadRequest, false},
		"Signup: Duplicate credentials":                      {"/signup", user.Username, user.Password, http.StatusUnprocessableEntity, false},
		"Signup: Different username with duplicate password": {"/signup", "NOTsimilar", user.Password, http.StatusOK, true},
		"Signup: Duplicate username with different password": {"/signup", user.Username, "NOTsimilar", http.StatusUnprocessableEntity, false},
		"Signup: Duplicate username with duplicate password": {"/signup", user.Username, user.Password, http.StatusUnprocessableEntity, false},
		"Login: Correct credentials":                         {"/login", user.Username, user.Password, http.StatusOK, true},
		"Login: Wrong username":                              {"/login", "NOTsimilar", user.Password, http.StatusUnauthorized, false},
		"Login: Wrong password":                              {"/login", user.Username, "NOTsimilar", http.StatusUnauthorized, false},
		"Login: Missing username":                            {"/login", "", user.Password, http.StatusBadRequest, false},
		"Login: Missing password":                            {"/login", user.Username, "", http.StatusBadRequest, false},
	}
	for name, test := range cases {
		db := newTestDB()
		if _, err := db.SignUp(user.Username, user.Password); err != nil {
			t.Errorf("%v: Failed to run test due to failing signup existing user: %s\n", name, err)
		}
		server, _ := testServer(t, db)
		response := testhelpers.TestRequest(t, server, http.MethodPost, test.path, "", handlers.NewClientCredentials(test.username, test.password))
		if response.Code != test.expectedStatusCode {
			t.Errorf("%v: Expects statuscode: %v but got: %v\n", name, test.expectedStatusCode, response.Code)
		}
		if result := testValidateToken(response.Body.String()); result != test.expectedValidToken {
			t.Errorf("%v: Expects valid token: %v but got: %v\n", name, test.expectedValidToken, result)
		}
	}
}

func TestHandlers_sendMessage(t *testing.T) {
	user := handlers.NewClientCredentials("ExistTest", "Passexist")
	cases := map[string]struct {
		path, username, password string
		topic                    string
		expectedFoundTopic       bool
	}{
		"Signup: Complete credentials":  {"/signup", "TestComplete", "Password", "newUser", true},
		"Signup: Duplicate credentials": {"/signup", user.Username, user.Password, "newUser", false},
		"Login: Correct credentials":    {"/login", user.Username, user.Password, "login", true},
		"Login: Wrong username":         {"/login", "NOTsimilar", user.Password, "login", false},
		"Login: Wrong password":         {"/login", user.Username, "NOTsimilar", "login", false},
	}
	for name, test := range cases {
		db := newTestDB()
		if _, err := db.SignUp(user.Username, user.Password); err != nil {
			t.Errorf("%v: Failed to run test due to failing signup existing user: %s\n", name, err)
			continue
		}
		server, channel := testServer(t, db)
		testhelpers.TestRequest(t, server, http.MethodPost, test.path, "", handlers.NewClientCredentials(test.username, test.password))
		go func() {
			time.Sleep(time.Millisecond)
			close(channel)
		}()
		foundTopic := false
		for message := range channel {
			if message.Topic != test.topic {
				t.Errorf("%v: Expects topic %v but got: %v\n", name, test.topic, message.Topic)
			} else {
				foundTopic = true
				var result types.Credentials
				if err := convert.FromJSONBytes(&result, []byte(message.Message)); err != nil {
					t.Errorf("%v: Expects to send a message containing an user but deserializing results in: %v\n", name, err)
				}
				if result.ID == 0 {
					t.Errorf("%v: Expects message with id but got id=0\n", name)
				}
				if result.Username != test.username {
					t.Errorf("%v: Expects message with username %v but got: %v\n", name, test.username, result.Username)
				}
			}
		}
		if foundTopic != test.expectedFoundTopic {
			t.Errorf("%v: Expects to found topic %v but got: %v\n", name, test.expectedFoundTopic, foundTopic)
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
	for name, test := range cases {
		db := newTestDB()
		server, _ := testServer(t, db)
		testhelpers.TestRequest(t, server, http.MethodPost, "/signup", "", handlers.NewClientCredentials(test.username, test.password))
		if _, ok := db.db[test.username]; ok != test.expected {
			t.Errorf("%v: Signup with username: %v and password: %v expects saving in DB: %v but got: %v\n", name, test.username, test.password, test.expected, ok)
		}
	}
}

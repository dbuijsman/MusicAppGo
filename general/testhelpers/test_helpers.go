package testhelpers

import (
	"context"
	"fmt"
	"general/convert"
	"general/server"
	"general/types"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
)

// TestWriter is an empty struct that can be used as an empty io.writer
type TestWriter struct{}

func (fake TestWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil
	return
}

// TestEmptyLogger return an empty logger
func TestEmptyLogger() *log.Logger {
	return log.New(TestWriter{}, "TEST", log.LstdFlags|log.Lshortfile)
}

// TestRequest sends a request to the server and returns the response. If token is not a default string, then it will add the token to the request
func TestRequest(t *testing.T, server *http.Server, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	bodyRequest, writer := io.Pipe()
	go func() {
		err := convert.WriteToJSON(body, writer)
		if err != nil {
			t.Fatalf("Error in test helper: %s", err)
		}
		writer.Close()
	}()
	request := httptest.NewRequest(method, "http://localhost"+path, bodyRequest)
	if token != "" {
		request.Header.Add("Token", token)
	}
	recorder := httptest.NewRecorder()
	server.Handler.ServeHTTP(recorder, request)
	return recorder
}

// Message is the type for testing messages to kafka
type Message struct {
	Topic   string
	Message []byte
}

// TestSendMessage returns a function that can be used for sending messages and a channel through which the message can be received
func TestSendMessage() (sendMessage func(string, []byte) error, channelReceiving chan Message) {
	channel := make(chan Message)
	sendMessage = func(topic string, message []byte) error {
		channel <- Message{Topic: topic, Message: message}
		return nil
	}
	channelReceiving = channel
	return
}

// THE REST IS OBSOLETE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

// TestSendMessageEmpty returns a SendMessageFunction for a handler that just returns nil
func TestSendMessageEmpty() func(topic string, message []byte) error {
	return func(topic string, message []byte) error {
		return nil
	}
}

// TestSendMessageWG returns a SendMessageFunction for a handler and two pointer to the topic and the message of the last message
func TestSendMessageWG(wg *sync.WaitGroup) (*string, *string, func(topic string, message []byte)) {
	topicValue, msgValue := "", ""
	top, msg := &topicValue, &msgValue
	return top, msg, func(topic string, message []byte) {
		*top = topic
		*msg = string(message)
		wg.Done()
	}
}

// TestSendMessageToParticularTopic returns a pointer to the message and a SendMessageFunction. It changes the value of the pointer only when the topic coincide with the given topic.
func TestSendMessageToParticularTopic(wg *sync.WaitGroup, specificTopic string) (*string, func(topic string, message []byte)) {
	msgValue := ""
	msg := &msgValue
	return msg, func(topic string, message []byte) {
		if topic != specificTopic {
			return
		}
		*msg = string(message)
		if wg != nil {
			wg.Done()
		}
	}
}

// TestPostRequest sends a post request to the given handler
func TestPostRequest(t *testing.T, handler func(http.ResponseWriter, *http.Request), body interface{}) *httptest.ResponseRecorder {
	bodyRequest, writer := io.Pipe()
	go func() {
		err := convert.WriteToJSON(body, writer)
		if err != nil {
			t.Fatalf("Error in test helper: %s", err)
		}
		writer.Close()
	}()
	request := httptest.NewRequest("POST", "/", bodyRequest)
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

// TestPostRequestWithContext sends a post request with the given context to the given handler
func TestPostRequestWithContext(t *testing.T, handler func(http.ResponseWriter, *http.Request), body interface{}, contextType, contextValue interface{}) *httptest.ResponseRecorder {
	bodyRequest, writer := io.Pipe()
	go func() {
		err := convert.WriteToJSON(body, writer)
		if err != nil {
			t.Fatalf("Error in test helper: %s", err)
		}
		writer.Close()
	}()
	request := httptest.NewRequest("POST", "/", bodyRequest)
	ctx := context.WithValue(request.Context(), contextType, contextValue)
	request = request.WithContext(ctx)
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

// TestGetRequestWithContext sends a get request to the given handler containing the given contextType and contextValue
func TestGetRequestWithContext(t *testing.T, handler func(http.ResponseWriter, *http.Request), contextType, contextValue interface{}) *httptest.ResponseRecorder {
	request := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(request.Context(), contextType, contextValue)
	request = request.WithContext(ctx)
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

// TestGetRequestWithPath sends a get request with a path variable to the given handler
func TestGetRequestWithPath(t *testing.T, handler func(http.ResponseWriter, *http.Request), pathVariable, pathValue, query string, middleware ...func(*log.Logger) func(http.Handler) http.Handler) *httptest.ResponseRecorder {
	if pathVariable == "" {
		t.Fatalf("Can't send get request due to empty pathVariable\n")
	}
	path := fmt.Sprintf("/{%v}", pathVariable)
	router := mux.NewRouter()
	emptyLogger := TestEmptyLogger()
	for _, middlewareFunction := range middleware {
		router.Use(middlewareFunction(emptyLogger))
	}
	router.Path(path).HandlerFunc(handler)
	var url string
	if query != "" {
		url = fmt.Sprintf("/%v?%v", pathValue, query)
	} else {
		url = fmt.Sprintf("/%v", pathValue)
	}
	url = strings.ReplaceAll(url, " ", "%20")
	request := httptest.NewRequest("GET", url, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// TestSendRequest sends the request to the given handler
func TestSendRequest(t *testing.T, handler func(http.ResponseWriter, *http.Request), request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	return recorder
}

// TestSendRequestWithPath sends a get request with a path variable to the given handler
func TestSendRequestWithPath(t *testing.T, handler func(http.ResponseWriter, *http.Request), pathVariable, pathValue string, request *http.Request) *httptest.ResponseRecorder {
	if pathVariable == "" {
		t.Fatalf("Can't send get request due to empty pathVariable\n")
	}
	path := fmt.Sprintf("/{%v}", pathVariable)
	router := mux.NewRouter()
	router.Path(path).HandlerFunc(handler)
	url := strings.ReplaceAll(fmt.Sprintf("/%v", pathValue), " ", "%20")
	request.URL.Path = url
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// WithCredentials add the credentials to the request
func WithCredentials(request *http.Request, creds types.Credentials) *http.Request {
	ctx := context.WithValue(request.Context(), types.Credentials{}, creds)
	return request.WithContext(ctx)
}

// WithOffsetMax add the offset and max value to the request
func WithOffsetMax(request *http.Request, offset, max int) *http.Request {
	ctx := context.WithValue(request.Context(), server.OffsetMax{}, server.OffsetMax{Offset: offset, Max: max})
	return request.WithContext(ctx)
}

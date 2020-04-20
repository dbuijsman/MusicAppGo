package common

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// TestPostRequest sends a post request to the given handler
func TestPostRequest(t *testing.T, handler func(http.ResponseWriter, *http.Request), body interface{}) *httptest.ResponseRecorder {
	bodyRequest, writer := io.Pipe()
	go func() {
		err := ToJSON(body, writer)
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

// TestGetRequestWithPathVariable sends a get request with a path variable to the given handler
func TestGetRequestWithPathVariable(t *testing.T, handler func(http.ResponseWriter, *http.Request), pathVariable, pathValue, query string, middleware ...func(*log.Logger) func(http.Handler) http.Handler) *httptest.ResponseRecorder {
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

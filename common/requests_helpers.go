package common

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// GetOffsetMaxFromRequest extracts the offset and max from the request.
func GetOffsetMaxFromRequest(request *http.Request) (offset, max int, err error) {
	queries := request.URL.Query()
	query := queries.Get("offset")
	offset, err = strconv.Atoi(query)
	if query != "" && err != nil {
		errorMessage := fmt.Sprintf("Got request with invalid value for offset: %s", err)
		err = errors.New(errorMessage)
		return
	}
	query = queries.Get("max")
	max, err = strconv.Atoi(query)
	if query != "" && err != nil {
		errorMessage := fmt.Sprintf("Got request with invalid value for max: %s", err)
		err = errors.New(errorMessage)
		return
	}
	err = nil
	if query == "" {
		max = 20
	}
	return
}

package common

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/go-playground/validator"
)

// ToJSON serializes the given interface into a string based JSON format
func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

// ToJSONBytes serializes the given interface into a byte array and returns the results
func ToJSONBytes(i interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := ToJSON(i, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromJSON deserializes the object from JSON string
// in an io.Reader to the given interface
func FromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	err := d.Decode(i)
	if err != nil {
		return err
	}
	return Validate(i)
}

// Validate checks if a certain struct has the right form
func Validate(i interface{}) error {
	validate := validator.New()
	err := validate.Struct(i)
	return err
}

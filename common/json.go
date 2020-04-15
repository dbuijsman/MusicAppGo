package common

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator"
)

// ToJSON serializes the given interface into a string based JSON format
func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
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

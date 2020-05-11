package general

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator"
)

// WriteToJSON serializes the given interface into a string based JSON format and it will write this to the writer
func WriteToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

// ToJSONBytes serializes the given interface into a byte array and returns the results
func ToJSONBytes(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// ReadFromJSON deserializes the object from an JSON string in an io.Reader to the given interface
func ReadFromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	err := d.Decode(i)
	if err != nil {
		return err
	}
	return Validate(i)
}

// ReadFromJSONNoValidation does the same as ReadFromJSON but without validating the struct
func ReadFromJSONNoValidation(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}

// FromJSONBytes deserializes the object from a JSON byte slice to the given interface
func FromJSONBytes(i interface{}, bytes []byte) error {
	err := json.Unmarshal(bytes, i)
	if err != nil {
		return err
	}
	return Validate(i)
}

// Validate checks if a certain struct has the right form
func Validate(i interface{}) error {
	validate := validator.New()
	return validate.Struct(i)
}

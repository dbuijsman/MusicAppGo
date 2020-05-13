package convert

import (
	"io"

	"github.com/go-yaml/yaml"
)

// ReadFromYAMLNoValidation does the same as ReadFromJSON but without validating the struct
func ReadFromYAMLNoValidation(i interface{}, r io.Reader) error {
	d := yaml.NewDecoder(r)
	return d.Decode(i)
}

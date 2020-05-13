package env

import (
	"errors"
	"flag"
	"fmt"
	"general/convert"
	"log"
	"os"
	"strconv"
	"strings"
)

var envVariables []envVar

func init() {
	envVariables = make([]envVar, 0)
}

// Cfg is the type that will contain all the configurations given by a yml file
type Cfg struct {
	Config map[string]string `yaml: config`
}

type envVar struct {
	name         string
	value        interface{}
	varType      string
	required     bool
	defaultValue interface{}
	setValue     func(interface{}, string) error
	setDefault   func(interface{}, interface{})
	validate     func(interface{}, bool) error
}

// SetString sets the default value of the environment variable with the given name
func SetString(name string, required bool, defaultValue string) *string {
	pointer := new(string)
	envVariables = append(envVariables, envVar{
		name:         name,
		value:        pointer,
		varType:      "string",
		required:     required,
		defaultValue: defaultValue,
		setValue: func(variable interface{}, value string) error {
			*variable.(*string) = value
			return nil
		},
		setDefault: func(variable interface{}, def interface{}) {
			*variable.(*string) = def.(string)
		},
		validate: func(variable interface{}, required bool) error {
			if required && *variable.(*string) == "" {
				return errors.New("Required value is zero")
			}
			return nil
		},
	})
	return pointer
}

// SetInt sets the default value of the environment variable with the given name
func SetInt(name string, required bool, defaultValue int) *int {
	pointer := new(int)
	envVariables = append(envVariables, envVar{
		name:         name,
		value:        pointer,
		varType:      "int",
		required:     required,
		defaultValue: defaultValue,
		setValue: func(variable interface{}, value string) error {
			valueInt, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				variable = nil
				return err
			}
			*variable.(*int) = int(valueInt)
			return nil
		},
		setDefault: func(variable interface{}, def interface{}) {
			*variable.(*int) = def.(int)
		},
		validate: func(variable interface{}, required bool) error {
			if required && *variable.(*int) == 0 {
				return errors.New("Required value is zero")
			}
			return nil
		},
	})
	return pointer
}

func processEnvVar(e envVar) error {
	envValue := os.Getenv(e.name)
	if envValue == "" {
		e.setDefault(e.value, e.defaultValue)
		return nil
	}
	return e.setValue(e.value, envValue)
}

func parseFlags() (configPath string) {
	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.Parse()
	return configPath
}
func readConfigFile(path string) map[string]string {
	cc := Cfg{Config: make(map[string]string)}
	if path == "" {
		return cc.Config
	}
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %s\n", err)
	}
	defer file.Close()
	if err = convert.ReadFromYAMLNoValidation(&cc, file); err != nil {
		log.Fatalf("Failed to decode config file: %s\n", err)
	}
	return cc.Config
}

// Parse parses the env
func Parse() error {
	configFromFile := readConfigFile(parseFlags())
	errors := make([]string, 0)
	for _, variable := range envVariables {
		if err := processEnvVar(variable); err != nil {
			errors = append(errors, fmt.Sprintf("%v: Got invalid value for type %v: %s\n", variable.name, variable.varType, err))
			continue
		}
		cfg, ok := configFromFile[variable.name]
		if !ok {
			continue
		}
		if err := variable.setValue(variable.value, cfg); err != nil {
			errors = append(errors, fmt.Sprintf("%v: Got invalid value for type %v: %s\n", variable.name, variable.varType, err))
			continue
		}
		if err := variable.validate(variable.value, variable.required); err != nil {
			errors = append(errors, fmt.Sprintf("%v: %s\n", variable.name, err))
		}
	}

	if len(errors) > 0 {
		errString := strings.Join(errors, "\n")
		return fmt.Errorf(errString)
	}
	return nil
}

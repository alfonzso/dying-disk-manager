package config

import (
	"fmt"
	"io/ioutil"

	"github.com/go-playground/validator/v10"

	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

func ReadConf(filename string) (*DDMConfig, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &DDMConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}

	validate = validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}

	return c, err
}

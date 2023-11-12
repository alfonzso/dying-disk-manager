package config

import (
	"fmt"
	"io/ioutil"

	// "reflect"

	"github.com/go-playground/validator/v10"

	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

// type ConfigStruct struct {
// 	Conf struct {
// 		Hits      int64
// 		Time      int64
// 		CamelCase string `yaml:"camelCase"`
// 	}
// }
// type ConfigStruct struct {
// 	Common struct {
// 		Mount struct {
// 			Enabled       bool `yaml:"enabled"`
// 			AfterAppStart bool `yaml:"afterAppStart"`
// 			PeriodicCheck struct {
// 				Enabled bool   `yaml:"enabled"`
// 				Cron    string `yaml:"cron"`
// 			} `yaml:"periodicCheck"`
// 		} `yaml:"mount"`
// 		Test struct {
// 			Enabled bool   `yaml:"enabled"`
// 			Cron    string `yaml:"cron"`
// 		} `yaml:"test"`
// 		Repair struct {
// 			Enabled       bool   `yaml:"enabled"`
// 			Command       string `yaml:"command"`
// 			CommandBefore string `yaml:"commandBefore"`
// 			CommandAfter  string `yaml:"commandAfter"`
// 		} `yaml:"repair"`
// 	} `yaml:"common"`
// 	Disks []struct {
// 		Name  string `yaml:"name"`
// 		UUID  string `yaml:"uuid" validate:"required"`
// 		Mount struct {
// 			Enabled       bool   `yaml:"enabled"`
// 			Path          string `yaml:"path"`
// 			PeriodicCheck struct {
// 				Enabled bool   `yaml:"enabled"`
// 				Cron    string `yaml:"cron"`
// 			} `yaml:"periodicCheck"`
// 		} `yaml:"mount"`
// 		Test struct {
// 			Enabled bool   `yaml:"enabled"`
// 			Cron    string `yaml:"cron"`
// 		} `yaml:"test"`
// 		Repair struct {
// 			Enabled       bool   `yaml:"enabled"`
// 			DiskNumber    int    `yaml:"diskNumber"`
// 			Command       string `yaml:"command"`
// 			CommandBefore string `yaml:"commandBefore"`
// 			CommandAfter  string `yaml:"commandAfter"`
// 		} `yaml:"repair"`
// 	} `yaml:"disks"`
// }

// func assertType(obj interface{}, expectedType reflect.Type) bool {
// 	fmt.Println(reflect.TypeOf(obj))
// 	return reflect.TypeOf(obj) == expectedType
// }

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

	// validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
	// 	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	// 	if name == "-" {
	// 		return ""
	// 	}
	// 	return name
	// })

	// validate.RegisterStructValidation(UserStructLevelValidation, User{})

	// err = validate.RegisterValidation("gender_custom_validation", func(fl validator.FieldLevel) bool {
	// 	value := fl.Field().Interface().(Disks)
	// 	return value != "unknown"
	// })

	validate = validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}

	// if assertType(c, reflect.TypeOf(&ConfigStruct{})) {
	// 	// Do something with the parsed data
	// 	fmt.Printf("keeekekekekekekkeke")
	// } else {
	// 	fmt.Println("Unmarshaled object has a different type than the expected struct.")
	// }

	fmt.Println(c)
	// fmt.Println(c.Disks[0].UUID)
	// fmt.Println(c.Disks[0].Mount.Enabled)
	// fmt.Println(strings.Split(c.Disks[0].UUID, ""))
	// fmt.Println(reflect.TypeOf(c.Disks[0].UUID))

	return c, err
}

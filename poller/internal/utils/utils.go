package utils

import (
	"errors"
	"os"
	"reflect"
	"strconv"
)

// LoadConfig populates string or int64 struct fields which are tagged with "env"
func LoadConfig(config interface{}) error {
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tagValue, ok := typ.Field(i).Tag.Lookup("env")
		if !ok {
			continue
		}

		value, ok := os.LookupEnv(tagValue)
		if !ok {
			return errors.New(tagValue + " env var is unset")
		}

		if value != "" {
			if field.Kind() == reflect.String {
				field.SetString(value)
			} else if field.Kind() == reflect.Int64 {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				field.SetInt(int64(intValue))
			}
		}
	}

	return nil
}

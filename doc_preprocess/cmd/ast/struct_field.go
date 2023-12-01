package main

import (
	"reflect"
	"strings"
)

type Filterable interface{}

func getField(data Filterable, tag string) reflect.Value {

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	tags := strings.Split(tag, ".")
	matchTag := tags[0]
	// Iterate over the fields of the struct
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)

		// Check if the field has the specified JSON tag
		if tagValue, ok := fieldType.Tag.Lookup("json"); ok && tagValue == matchTag {
			if strings.Contains(tag, ".") {
				parts := strings.Join(tags[1:], ".")
				return getField(field.Interface(), parts)
			} else {
				return field
			}
		}
	}
	return reflect.Value{}
}

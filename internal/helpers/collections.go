package helpers

import (
	"fmt"
	"reflect"
	"sort"
)

func list(values ...any) []any {
	return values
}

func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires key/value pairs")
	}

	result := make(map[string]any, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		key, ok := values[index].(string)
		if !ok {
			return nil, fmt.Errorf("dict key at position %d must be a string", index)
		}
		result[key] = values[index+1]
	}
	return result, nil
}

func in(needle, collection any) bool {
	if collection == nil {
		return false
	}

	value := reflect.ValueOf(collection)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			if reflect.DeepEqual(needle, value.Index(index).Interface()) {
				return true
			}
		}
	case reflect.Map:
		needleValue := reflect.ValueOf(needle)
		if needleValue.IsValid() && needleValue.Type().AssignableTo(value.Type().Key()) {
			return value.MapIndex(needleValue).IsValid()
		}
	}

	return false
}

func keys(value any) ([]string, error) {
	reflection := reflect.ValueOf(value)
	if !reflection.IsValid() || reflection.Kind() != reflect.Map || reflection.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("keys requires a map with string keys")
	}

	result := make([]string, 0, reflection.Len())
	iterator := reflection.MapRange()
	for iterator.Next() {
		result = append(result, iterator.Key().String())
	}
	sort.Strings(result)
	return result, nil
}

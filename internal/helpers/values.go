package helpers

import "reflect"

func defaultValue(fallback, value any) any {
	if isEmpty(value) {
		return fallback
	}
	return value
}

func coalesce(values ...any) any {
	for _, value := range values {
		if !isEmpty(value) {
			return value
		}
	}
	return nil
}

func ternary(whenTrue, whenFalse any, condition bool) any {
	if condition {
		return whenTrue
	}
	return whenFalse
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	reflection := reflect.ValueOf(value)
	for reflection.Kind() == reflect.Interface || reflection.Kind() == reflect.Pointer {
		if reflection.IsNil() {
			return true
		}
		reflection = reflection.Elem()
	}

	switch reflection.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return reflection.Len() == 0
	case reflect.Bool:
		return !reflection.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflection.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return reflection.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return reflection.Float() == 0
	default:
		return reflection.IsZero()
	}
}

// Package validation validates structs against `validate` struct tags.
package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const tagName = "validate"

// Validate checks v against its `validate` struct tags and returns Errors
// when any rule fails, or nil when v is valid. v must be a struct or a
// pointer to one.
//
//	type RegisterInput struct {
//		Name     string `validate:"required"`
//		Email    string `validate:"required,email"`
//		Password string `validate:"required,min=8,max=1024"`
//		Confirm  string `validate:"required,eqfield=Password" label:"Password confirmation"`
//	}
//
//	if err := validation.Validate(input); err != nil {
//		var errs validation.Errors
//		errors.As(err, &errs)
//		return c.RenderStatus(http.StatusUnprocessableEntity, "register", route.Data{
//			"Errors": errs,
//		})
//	}
func Validate(v any) error {
	value := reflect.ValueOf(v)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("validation: nil pointer")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("validation: expected a struct, got %s", value.Kind())
	}

	errs := make(Errors)
	validateStruct(value, errs)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func validateStruct(value reflect.Value, errs Errors) {
	structType := value.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue := value.Field(i)
		label := fieldLabel(field)

		for rule := range strings.SplitSeq(tag, ",") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}
			ruleName, param, _ := strings.Cut(rule, "=")
			if message, ok := applyRule(ruleName, param, fieldValue, value); !ok {
				errs.Add(label, message)
			}
		}
	}
}

func fieldLabel(field reflect.StructField) string {
	if label := field.Tag.Get("label"); label != "" {
		return label
	}
	return field.Name
}

// applyRule reports (failure message, passed). The message is only used when passed is false.
func applyRule(rule, param string, field, parent reflect.Value) (string, bool) {
	switch rule {
	case "required":
		return "is required", !field.IsZero()
	case "email":
		return "must be a valid email address", field.IsZero() || isValidEmail(field.String())
	case "url":
		return "must be a valid URL", field.IsZero() || isValidURL(field.String())
	case "numeric":
		return "must contain only digits", field.IsZero() || isNumeric(field.String())
	case "alpha":
		return "must contain only letters", field.IsZero() || isAlpha(field.String())
	case "alphanumeric":
		return "must contain only letters and numbers", field.IsZero() || isAlphanumeric(field.String())
	case "min":
		return minRule(field, param)
	case "max":
		return maxRule(field, param)
	case "len":
		return lenRule(field, param)
	case "oneof":
		return oneofRule(field, param)
	case "eqfield":
		return eqfieldRule(field, param, parent)
	default:
		return fmt.Sprintf("has an unknown validation rule %q", rule), false
	}
}

func minRule(field reflect.Value, param string) (string, bool) {
	limit, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return fmt.Sprintf("has an invalid min rule %q", param), false
	}
	if field.IsZero() {
		return "", true
	}
	switch field.Kind() {
	case reflect.String:
		return fmt.Sprintf("must be at least %s characters", param), float64(len([]rune(field.String()))) >= limit
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("must be at least %s", param), float64(field.Int()) >= limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("must be at least %s", param), float64(field.Uint()) >= limit
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("must be at least %s", param), field.Float() >= limit
	case reflect.Slice, reflect.Array, reflect.Map:
		return fmt.Sprintf("must have at least %s items", param), float64(field.Len()) >= limit
	default:
		return fmt.Sprintf("min rule is not supported for %s", field.Kind()), false
	}
}

func maxRule(field reflect.Value, param string) (string, bool) {
	limit, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return fmt.Sprintf("has an invalid max rule %q", param), false
	}
	if field.IsZero() {
		return "", true
	}
	switch field.Kind() {
	case reflect.String:
		return fmt.Sprintf("must be at most %s characters", param), float64(len([]rune(field.String()))) <= limit
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("must be at most %s", param), float64(field.Int()) <= limit
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("must be at most %s", param), float64(field.Uint()) <= limit
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("must be at most %s", param), field.Float() <= limit
	case reflect.Slice, reflect.Array, reflect.Map:
		return fmt.Sprintf("must have at most %s items", param), float64(field.Len()) <= limit
	default:
		return fmt.Sprintf("max rule is not supported for %s", field.Kind()), false
	}
}

func lenRule(field reflect.Value, param string) (string, bool) {
	length, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Sprintf("has an invalid len rule %q", param), false
	}
	switch field.Kind() {
	case reflect.String:
		return fmt.Sprintf("must be exactly %s characters", param), len([]rune(field.String())) == length
	case reflect.Slice, reflect.Array, reflect.Map:
		return fmt.Sprintf("must have exactly %s items", param), field.Len() == length
	default:
		return fmt.Sprintf("len rule is not supported for %s", field.Kind()), false
	}
}

func oneofRule(field reflect.Value, param string) (string, bool) {
	options := strings.Split(param, "|")
	if field.IsZero() {
		return "", true
	}
	value := fmt.Sprintf("%v", field.Interface())
	if slices.Contains(options, value) {
		return "", true
	}
	return fmt.Sprintf("must be one of %s", strings.Join(options, ", ")), false
}

func eqfieldRule(field reflect.Value, param string, parent reflect.Value) (string, bool) {
	other := parent.FieldByName(param)
	if !other.IsValid() {
		return fmt.Sprintf("references an unknown field %q", param), false
	}
	return fmt.Sprintf("must match %s", param), reflect.DeepEqual(field.Interface(), other.Interface())
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

func isValidURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isAlpha(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !isLetter(r) {
			return false
		}
	}
	return true
}

func isAlphanumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !isLetter(r) && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

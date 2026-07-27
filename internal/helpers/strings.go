package helpers

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

func upper(value string) string {
	return strings.ToUpper(value)
}

func lower(value string) string {
	return strings.ToLower(value)
}

func trim(value string) string {
	return strings.TrimSpace(value)
}

func contains(part, value string) bool {
	return strings.Contains(value, part)
}

func hasPrefix(prefix, value string) bool {
	return strings.HasPrefix(value, prefix)
}

func hasSuffix(suffix, value string) bool {
	return strings.HasSuffix(value, suffix)
}

func trimPrefix(prefix, value string) string {
	return strings.TrimPrefix(value, prefix)
}

func trimSuffix(suffix, value string) string {
	return strings.TrimSuffix(value, suffix)
}

func replace(old, replacement, value string) string {
	return strings.ReplaceAll(value, old, replacement)
}

func split(separator, value string) []string {
	return strings.Split(value, separator)
}

func join(separator string, values []string) string {
	return strings.Join(values, separator)
}

func capitalize(value string) string {
	value = strings.TrimSpace(value)
	first, size := utf8.DecodeRuneInString(value)
	if first == utf8.RuneError && size == 0 {
		return value
	}
	return string(unicode.ToUpper(first)) + value[size:]
}

func slug(value string) string {
	var output strings.Builder
	pendingDash := false

	for _, character := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if pendingDash && output.Len() > 0 {
				output.WriteByte('-')
			}
			output.WriteRune(character)
			pendingDash = false
			continue
		}
		pendingDash = true
	}

	return output.String()
}

func truncate(length int, value string) string {
	if length <= 0 {
		return ""
	}

	characters := []rune(value)
	if len(characters) <= length {
		return value
	}
	if length <= 3 {
		return string(characters[:length])
	}
	return string(characters[:length-3]) + "..."
}

func queryEscape(value string) string {
	return url.QueryEscape(value)
}

package validation

import "strings"

// Errors maps a struct field's label to the validation messages that failed for it.
type Errors map[string][]string

// Error implements the error interface by joining every message.
func (e Errors) Error() string {
	var b strings.Builder
	first := true
	for field, messages := range e {
		for _, message := range messages {
			if !first {
				b.WriteString("; ")
			}
			first = false
			b.WriteString(field)
			b.WriteString(" ")
			b.WriteString(message)
		}
	}
	return b.String()
}

// Add appends a failure message for field.
func (e Errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Has reports whether field has any recorded failure.
func (e Errors) Has(field string) bool {
	return len(e[field]) > 0
}

// First returns the first recorded failure message for field, or "" when none.
func (e Errors) First(field string) string {
	messages := e[field]
	if len(messages) == 0 {
		return ""
	}
	return messages[0]
}

// Package helpers contains globally available framework template helpers.
package helpers

import (
	"html/template"
	"strings"
)

// TemplateFuncs returns a fresh registry of global template helpers.
// Add new framework-wide helpers to this map.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"default": func(fallback, value string) string {
			if strings.TrimSpace(value) == "" {
				return fallback
			}
			return value
		},
	}
}

// Package helpers contains globally available framework template helpers.
package helpers

import (
	"html/template"
)

// TemplateFuncs returns a fresh registry of global template helpers.
// Add new framework-wide helpers to this map.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Strings
		"upper":       upper,
		"lower":       lower,
		"trim":        trim,
		"contains":    contains,
		"hasPrefix":   hasPrefix,
		"hasSuffix":   hasSuffix,
		"trimPrefix":  trimPrefix,
		"trimSuffix":  trimSuffix,
		"replace":     replace,
		"split":       split,
		"join":        join,
		"capitalize":  capitalize,
		"slug":        slug,
		"truncate":    truncate,
		"queryEscape": queryEscape,
		"csrfField":   csrfField,

		// Values and conditions
		"default":  defaultValue,
		"coalesce": coalesce,
		"ternary":  ternary,

		// Collections
		"list": list,
		"dict": dict,
		"in":   in,
		"keys": keys,

		// Numbers
		"add":  add,
		"sub":  sub,
		"mul":  mul,
		"div":  divide,
		"mod":  modulo,
		"inc":  increment,
		"dec":  decrement,
		"even": even,
		"odd":  odd,
		"seq":  sequence,

		// Date and time
		"date": formatDate,
		"now":  now,
	}
}

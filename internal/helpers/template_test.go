package helpers_test

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	"fluxgo/internal/helpers"
)

func TestGlobalTemplateHelpers(t *testing.T) {
	source := strings.Join([]string{
		`{{upper "flux"}}`,
		`{{"  GO  " | trim | lower}}`,
		`{{default "fallback" ""}}`,
		`{{true | ternary "yes" "no"}}`,
		`{{join "," (split " " "a b")}}`,
		`{{slug "Hello, Go World!"}}`,
		`{{truncate 7 "abcdefghij"}}`,
		`{{add 2 3 4}}`,
		`{{range seq 1 3}}{{.}}{{end}}`,
		`{{if in "go" (list "html" "go")}}found{{end}}`,
		`{{date "2006-01-02" .Date}}`,
		`{{queryEscape "go html"}}`,
	}, "|")

	compiled, err := template.New("helpers").Funcs(helpers.TemplateFuncs()).Parse(source)
	if err != nil {
		t.Fatalf("parse helpers: %v", err)
	}

	var output bytes.Buffer
	err = compiled.Execute(&output, map[string]any{
		"Date": time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("execute helpers: %v", err)
	}

	want := "FLUX|go|fallback|yes|a,b|hello-go-world|abcd...|9|123|found|2026-07-27|go&#43;html"
	if output.String() != want {
		t.Fatalf("expected %q, got %q", want, output.String())
	}
}

func TestInvalidHelperInputReturnsExecutionError(t *testing.T) {
	compiled := template.Must(
		template.New("invalid").Funcs(helpers.TemplateFuncs()).Parse(`{{div 10 0}}`),
	)

	if err := compiled.Execute(&bytes.Buffer{}, nil); err == nil {
		t.Fatal("expected division by zero to return an execution error")
	}
}

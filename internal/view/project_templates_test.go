package view_test

import (
	"testing"

	"fluxgo/internal/view"
)

func TestProjectTemplatesCompile(t *testing.T) {
	if _, err := view.New(view.Config{Root: "../../views"}); err != nil {
		t.Fatalf("compile project templates: %v", err)
	}
}

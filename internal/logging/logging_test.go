package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fluxgo/internal/logging"
	Route "fluxgo/internal/route"
)

func TestNewProducesJSONWhenConfigured(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Config{Format: "json", Output: &buf})
	logger.Info("hello", "key", "value")

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON output, got %q: %v", buf.String(), err)
	}
	if decoded["msg"] != "hello" || decoded["key"] != "value" {
		t.Fatalf("unexpected log record: %v", decoded)
	}
}

func TestNewRespectsConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.New(logging.Config{Level: "error", Format: "text", Output: &buf})

	logger.Info("should be filtered out")
	if buf.Len() != 0 {
		t.Fatalf("expected info logs to be filtered at error level, got %q", buf.String())
	}

	logger.Error("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Fatalf("expected the error log to be written, got %q", buf.String())
	}
}

func TestMiddlewareLogsRequestOutcome(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	engine := Route.New()
	engine.Use(logging.Middleware(logger))
	engine.Get("/ping", func(c *Route.Context) error {
		return c.OK(Route.Data{"pong": true})
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ping", nil))

	output := buf.String()
	if !strings.Contains(output, "method=GET") || !strings.Contains(output, "path=/ping") ||
		!strings.Contains(output, "status=200") {
		t.Fatalf("expected structured request log, got %q", output)
	}
}

func TestMiddlewareLogsErrorsAtErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	engine := Route.New()
	engine.Use(logging.Middleware(logger))
	engine.Get("/fail", func(c *Route.Context) error {
		return Route.BadRequestError("nope")
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/fail", nil))

	if !strings.Contains(buf.String(), "level=ERROR") {
		t.Fatalf("expected an error-level log entry, got %q", buf.String())
	}
}

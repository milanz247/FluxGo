package middleware_test

import (
	"reflect"
	"testing"

	"fluxgo/internal/middleware"
)

func TestEngineWrapsInRegistrationOrder(t *testing.T) {
	type handler func()

	var calls []string
	engine := middleware.New[handler]()
	engine.Use(
		func(next handler) handler {
			return func() {
				calls = append(calls, "first")
				next()
			}
		},
		func(next handler) handler {
			return func() {
				calls = append(calls, "second")
				next()
			}
		},
	)

	engine.Wrap(func() {
		calls = append(calls, "handler")
	})()

	want := []string{"first", "second", "handler"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("expected calls %v, got %v", want, calls)
	}
}

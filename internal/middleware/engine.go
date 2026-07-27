// Package middleware provides the middleware chain used by the framework.
package middleware

// Middleware wraps a handler with additional behaviour.
type Middleware[T any] func(T) T

// Engine stores and applies an ordered middleware chain.
type Engine[T any] struct {
	middleware []Middleware[T]
}

// New creates an empty middleware engine.
func New[T any]() *Engine[T] {
	return &Engine[T]{}
}

// Use appends middleware to the chain.
func (e *Engine[T]) Use(middleware ...Middleware[T]) {
	e.middleware = append(e.middleware, middleware...)
}

// Wrap applies the chain so middleware execute in registration order.
func (e *Engine[T]) Wrap(handler T) T {
	for i := len(e.middleware) - 1; i >= 0; i-- {
		handler = e.middleware[i](handler)
	}
	return handler
}

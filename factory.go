package quarry

import (
	"context"
	"sync"
)

// Dependencies is a map of named values that are provided to Factories.
type Dependencies map[string]interface{}

// Contains returns true when these dependencies has the given key.
func (d Dependencies) Contains(key string) bool {
	_, ok := d[key]
	return ok
}

// Factory is a function that executes and creates a result.
type Factory func(ctx context.Context, params interface{}, deps Dependencies) (interface{}, error)

// Provider creates a Factory that returns a value.
func Provider(value interface{}) Factory {
	return func(context.Context, interface{}, Dependencies) (interface{}, error) {
		return value, nil
	}
}

// Singleton wraps a Factory-like function to ensure that it is used only once.
// Unlike Factory, the function does not use parameters.
// Its return value will be re-used, including errors.
func Singleton(factory func(ctx context.Context, deps Dependencies) (interface{}, error)) Factory {
	var once sync.Once
	var result interface{}
	var err error
	return func(ctx context.Context, params interface{}, deps Dependencies) (interface{}, error) {
		once.Do(func() {
			result, err = factory(ctx, deps)
		})
		return result, err
	}
}

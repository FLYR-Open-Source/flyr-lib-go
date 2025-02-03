package context

import (
	"context"
	"fmt"
)

// ContextKey is the standard type used for specifying the key value used to add and retrieve objects from the context.
type ContextKey string

// GetObjectFromContext retrieves an object using the given ContextKey, casts it to the specified type, and returns it.
// An error is returned if no object is found for the given key or it cannot be cast to the specified type.
func GetObjectFromContext[T any](ctx context.Context, key ContextKey) (T, error) {
	object, ok := ctx.Value(key).(T)
	if !ok {
		return object, fmt.Errorf("context object \"%s\" was not found or is not the expected type", key)
	}

	return object, nil
}

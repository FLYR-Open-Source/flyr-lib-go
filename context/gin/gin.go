package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetObjectFromGinContext retrieves an object using the given context key, casts it to the specified type, and returns it.
//
// An error is returned if no object is found for the given key or it cannot be cast to the specified type.
func GetObjectFromGinContext[T any](ctx *gin.Context, key string) (T, error) {
	object, ok := ctx.Get(key)
	if !ok {
		return *new(T), fmt.Errorf("context object with key %q was not found", key)
	}

	result, ok := object.(T)
	if !ok {
		return *new(T), fmt.Errorf("context object with key %q is not the expected type", key)
	}

	return result, nil
}

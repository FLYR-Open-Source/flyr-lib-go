package context_test

import (
	"context"
	"testing"

	flyrContext "github.com/FlyrInc/flyr-lib-go/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock struct and interface

type MockHelpersBaseInterface interface {
	TestFunction()
}

type MockHelpersBaseStruct struct{}

func (mock *MockHelpersBaseStruct) TestFunction() {}

// Test setup

func getTestHelpersBaseContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, flyrContext.ContextKey("mockStruct"), &MockHelpersBaseStruct{})

	return ctx
}

// Tests

func TestGetObjectFromContextReturnsExpectedOutput(t *testing.T) {
	ctx := getTestHelpersBaseContext()

	object, err := flyrContext.GetObjectFromContext[MockHelpersBaseInterface](ctx, flyrContext.ContextKey("mockStruct"))

	require.NoError(t, err)

	assert.NotNil(t, object)
	assert.IsType(t, &MockHelpersBaseStruct{}, object)
}

func TestGetObjectFromContextReturnsErrorOnMissingObject(t *testing.T) {
	ctx := getTestHelpersBaseContext()

	_, err := flyrContext.GetObjectFromContext[MockHelpersBaseInterface](ctx, flyrContext.ContextKey("test"))

	assert.Error(t, err)
}

func TestGetObjectFromContextReturnsErrorOnInvalidType(t *testing.T) {
	ctx := getTestHelpersBaseContext()

	_, err := flyrContext.GetObjectFromContext[string](ctx, flyrContext.ContextKey("mockStruct"))

	assert.Error(t, err)
}

package base_test

import (
	"context"
	"testing"

	flyrContextBase "github.com/FlyrInc/flyr-lib-go/context/base"

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

func getTestHelpersBaseContext() (context.Context, *MockHelpersBaseStruct) {
	ctx := context.Background()
	object := &MockHelpersBaseStruct{}

	ctx = context.WithValue(ctx, flyrContextBase.ContextKey("mockStruct"), object)

	return ctx, object
}

// Tests

func TestGetObjectFromContextReturnsExpectedOutput(t *testing.T) {
	ctx, originalObject := getTestHelpersBaseContext()

	object, err := flyrContextBase.GetObjectFromContext[MockHelpersBaseInterface](ctx, flyrContextBase.ContextKey("mockStruct"))

	require.NoError(t, err)
	require.IsType(t, &MockHelpersBaseStruct{}, object)

	assert.Equal(t, originalObject, object)
}

func TestGetObjectFromContextReturnsErrorOnMissingObject(t *testing.T) {
	ctx, _ := getTestHelpersBaseContext()

	_, err := flyrContextBase.GetObjectFromContext[MockHelpersBaseInterface](ctx, flyrContextBase.ContextKey("test"))

	assert.Error(t, err)
}

func TestGetObjectFromContextReturnsErrorOnInvalidType(t *testing.T) {
	ctx, _ := getTestHelpersBaseContext()

	_, err := flyrContextBase.GetObjectFromContext[string](ctx, flyrContextBase.ContextKey("mockStruct"))

	assert.Error(t, err)
}

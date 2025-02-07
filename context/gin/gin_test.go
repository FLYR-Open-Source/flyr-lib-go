package context_test

import (
	"testing"

	flyrContextGin "github.com/FlyrInc/flyr-lib-go/context/gin"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock struct and interface

type MockHelpersGinInterface interface {
	TestFunction()
}

type MockHelpersGinStruct struct{}

func (mock *MockHelpersGinStruct) TestFunction() {}

// Test setup

func getTestHelpersGinContext() (*gin.Context, *MockHelpersGinStruct) {
	ctx := &gin.Context{}
	object := &MockHelpersGinStruct{}
	ctx.Set("mockStruct", object)
	return ctx, object
}

// Tests

func TestGetObjectFromGinContextReturnsExpectedOutput(t *testing.T) {
	ctx, originalObject := getTestHelpersGinContext()

	object, err := flyrContextGin.GetObjectFromGinContext[MockHelpersGinInterface](ctx, "mockStruct")

	require.NoError(t, err)
	require.IsType(t, &MockHelpersGinStruct{}, object)

	assert.Equal(t, originalObject, object)
}

func TestGetObjectFromGinContextReturnsErrorOnMissingObject(t *testing.T) {
	ctx, _ := getTestHelpersGinContext()

	_, err := flyrContextGin.GetObjectFromGinContext[MockHelpersGinInterface](ctx, "test")

	assert.Error(t, err)
}

func TestGetObjectFromGinContextReturnsErrorOnInvalidType(t *testing.T) {
	ctx, _ := getTestHelpersGinContext()

	_, err := flyrContextGin.GetObjectFromGinContext[string](ctx, "mockStruct")

	assert.Error(t, err)
}

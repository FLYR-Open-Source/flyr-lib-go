package flyr_context_test

import (
	"testing"

	flyrContext "github.com/FlyrInc/flyr-lib-go/context"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock struct and interface

type MockHelpersGinInterface interface {
	TestFunction()
}

type MockHelpersGinStruct struct{}

func (mock *MockHelpersGinStruct) TestFunction() {}

// Test setup

func getTestHelpersGinContext() *gin.Context {
	ctx := &gin.Context{}
	ctx.Set("mockStruct", &MockHelpersGinStruct{})
	return ctx
}

// Tests

func TestGetObjectFromGinContextReturnsExpectedOutput(t *testing.T) {
	ctx := getTestHelpersGinContext()

	object, err := flyrContext.GetObjectFromGinContext[MockHelpersGinInterface](ctx, "mockStruct")

	assert.NoError(t, err)

	assert.NotNil(t, object)
	assert.IsType(t, &MockHelpersGinStruct{}, object)
}

func TestGetObjectFromGinContextReturnsErrorOnMissingObject(t *testing.T) {
	ctx := getTestHelpersGinContext()

	_, err := flyrContext.GetObjectFromGinContext[MockHelpersGinInterface](ctx, "test")

	assert.Error(t, err)
}

func TestGetObjectFromGinContextReturnsErrorOnInvalidType(t *testing.T) {
	ctx := getTestHelpersGinContext()

	_, err := flyrContext.GetObjectFromGinContext[string](ctx, "mockStruct")

	assert.Error(t, err)
}

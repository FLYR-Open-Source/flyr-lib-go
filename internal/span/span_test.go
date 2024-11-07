package span

import (
	"context"
	"errors"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
)

func TestEndWithError(t *testing.T) {
	t.Run("With Error", func(t *testing.T) {
		_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
		defer fakeSpan.End()
		err := errors.New("test error")

		span := Span{Span: &fakeSpan}
		span.EndWithError(err)

		require.ErrorIs(t, err, fakeSpan.FakeRecordedError.Error)
		assert.Equal(t, codes.Error, fakeSpan.FakeStatus.Code)
		assert.Equal(t, fakeSpan.FakeStatus.Description, err.Error())
	})

	t.Run("Without Error", func(t *testing.T) {
		_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
		defer fakeSpan.End()

		span := Span{Span: &fakeSpan}
		span.EndWithError(nil)

		require.NoError(t, fakeSpan.FakeRecordedError.Error)
		assert.Equal(t, codes.Unset, fakeSpan.FakeStatus.Code)
		assert.Equal(t, "", fakeSpan.FakeStatus.Description)
	})
}

func TestEndSuccessfully(t *testing.T) {
	_, fakeSpan := testhelpers.GetFakeSpan(context.Background())
	defer fakeSpan.End()

	span := Span{Span: &fakeSpan}
	span.EndSuccessfully()

	require.NoError(t, fakeSpan.FakeRecordedError.Error)
	assert.Equal(t, codes.Ok, fakeSpan.FakeStatus.Code)
	assert.Equal(t, "", fakeSpan.FakeStatus.Description)
}

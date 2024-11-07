package logger

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/FlyrInc/flyr-lib-go/config"
	"github.com/stretchr/testify/assert"
)

func TestGetAttributes(t *testing.T) {
	args := []interface{}{"key1", "value1", "key2", "value2"}

	t.Run("With correct code details", func(t *testing.T) {
		attrs := getAttributes(context.Background(), nil, args...)

		assert.GreaterOrEqual(t, len(attrs), 4)

		codePath := attrs[0].(slog.Attr)
		assert.Equal(t, config.CODE_PATH, codePath.Key)
		assert.Contains(t, codePath.Value.String(), "src/testing/testing.go")

		codeLine := attrs[1].(slog.Attr)
		assert.Equal(t, config.CODE_LINE, codeLine.Key)
		assert.Positive(t, codeLine.Value.Int64())

		codeFunc := attrs[2].(slog.Attr)
		assert.Equal(t, config.CODE_FUNC, codeFunc.Key)
		assert.Contains(t, codeFunc.Value.String(), "tRunner")

		codeNs := attrs[3].(slog.Attr)
		assert.Equal(t, config.CODE_NS, codeNs.Key)
		assert.Contains(t, codeNs.Value.String(), "testing")
	})

	t.Run("With metadata", func(t *testing.T) {
		attrs := getAttributes(context.Background(), nil, args...)

		assert.Len(t, attrs, 5)

		metadata := attrs[4].(slog.Attr)
		assert.Equal(t, config.LOG_METADATA_KEY, metadata.Key)
		assert.Equal(t, "[key1=value1 key2=value2]", metadata.Value.String())
	})

	t.Run("With an error", func(t *testing.T) {
		err := errors.New("test error")
		attrs := getAttributes(context.Background(), err, args...)

		assert.Len(t, attrs, 6)

		errorMessage := attrs[5].(slog.Attr)
		assert.Equal(t, config.LOG_ERROR_KEY, errorMessage.Key)
		assert.Equal(t, err.Error(), errorMessage.Value.String())
	})

	t.Run("Without extra metadata", func(t *testing.T) {
		attrs := getAttributes(context.Background(), nil)

		assert.Len(t, attrs, 5)

		metadata := attrs[4].(slog.Attr)
		assert.Equal(t, config.LOG_METADATA_KEY, metadata.Key)
		assert.Equal(t, "[]", metadata.Value.String())
	})
}

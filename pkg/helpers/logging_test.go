package helpers

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrAttr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected slog.Attr
	}{
		{
			name:     "nil error should return empty slog.Attr",
			err:      nil,
			expected: slog.Attr{},
		},
		{
			name:     "error message should be set as value with error key",
			err:      errors.New("error message"),
			expected: slog.Any("error", errors.New("error message")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := ErrAttr(test.err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func ExampleErrAttr() {
	err := errors.New("example error")
	attr := ErrAttr(err)
	slog.Info("example error", attr)
}

package compressor

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressor(t *testing.T) {
	logger := slog.Default()
	compressor := NewCompressor(logger)

	tests := []struct {
		name        string
		input       []byte
		expected    []byte
		expectError bool
	}{
		{
			name:        "Compress and Decompress valid data",
			input:       []byte("Hello, World!"),
			expectError: false,
		},
		{
			name:        "Compress empty data",
			input:       []byte(""),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Compression
			compressedData, err := compressor.Compress(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test Decompression
			decompressedData, err := compressor.Decompress(bytes.NewReader(compressedData))
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decompressedData)
		})
	}

	// Test Decompression with invalid data
	t.Run("Decompress invalid data", func(t *testing.T) {
		invalidData := []byte("invalid compressed data")
		_, err := compressor.Decompress(bytes.NewReader(invalidData))
		assert.Error(t, err)
	})
}

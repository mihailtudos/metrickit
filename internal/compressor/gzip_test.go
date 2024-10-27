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

func BenchmarkCompressor_Compress(b *testing.B) {
	logger := slog.Default()
	compressor := NewCompressor(logger)

	input := []byte("Hello, World!") // Change this to larger data if needed

	for i := 0; i < b.N; i++ {
		_, err := compressor.Compress(input)
		if err != nil {
			b.Fatalf("Compress failed: %v", err)
		}
	}
}

func BenchmarkCompressor_Decompress(b *testing.B) {
	logger := slog.Default()
	compressor := NewCompressor(logger)

	// Compress some data to have valid input for decompression
	input := []byte("Hello, World!")
	compressedData, err := compressor.Compress(input)
	if err != nil {
		b.Fatalf("Compress failed: %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := compressor.Decompress(bytes.NewReader(compressedData))
		if err != nil {
			b.Fatalf("Decompress failed: %v", err)
		}
	}
}

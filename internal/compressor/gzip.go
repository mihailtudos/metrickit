// Package compressor provides functionality for compressing and decompressing
// data using the gzip format.
//
// It includes methods for compressing data into a gzip format and decompressing
// data from a gzip format. The package is useful for reducing the size of data
// being stored or transmitted and for handling compression-related errors.
package compressor

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// Compressor is a struct that provides methods to compress and decompress data
// with gzip. It also includes logging capabilities to handle errors.
type Compressor struct {
	logger *slog.Logger // Logger for recording error messages.
}

// NewCompressor creates a new Compressor instance with the provided logger.
//
// Parameters:
//   - logger: The logger to use for error reporting.
//
// Returns:
//   - Compressor: A new Compressor instance.
func NewCompressor(logger *slog.Logger) Compressor {
	return Compressor{
		logger: logger,
	}
}

// Compress compresses the input data using gzip and returns the compressed
// byte slice.
//
// Parameters:
//   - data: The byte slice containing the data to be compressed.
//
// Returns:
//   - []byte: The compressed data in gzip format.
//   - error: An error if the compression process fails.
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	// Write data to the gzip writer.
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	// Close the gzip writer to flush any remaining data.
	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Decompress reads data from a gzip reader and returns the decompressed byte slice.
//
// Parameters:
//   - r: An io.Reader from which the compressed data is read.
//
// Returns:
//   - []byte: The decompressed data.
//   - error: An error if the decompression process fails.
func (c *Compressor) Decompress(r io.Reader) ([]byte, error) {
	// Create a new gzip reader.
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	// Ensure the gzip reader is closed properly, logging any error if it occurs.
	defer func() {
		if err := gr.Close(); err != nil {
			c.logger.ErrorContext(
				context.Background(),
				"failed to close the reader",
				helpers.ErrAttr(err))
		}
	}()

	var b bytes.Buffer
	// Read decompressed data into a buffer.
	if _, err := b.ReadFrom(gr); err != nil {
		return nil, fmt.Errorf("decompress function failed to read from the writer: %w", err)
	}

	return b.Bytes(), nil
}

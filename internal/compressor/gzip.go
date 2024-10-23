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

type Compressor struct {
	logger *slog.Logger
}

func NewCompressor(logger *slog.Logger) Compressor {
	return Compressor{
		logger: logger,
	}
}

func (c *Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (c *Compressor) Decompress(r io.Reader) ([]byte, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() {
		if err := gr.Close(); err != nil {
			c.logger.ErrorContext(
				context.Background(),
				"failed to close the reader",
				helpers.ErrAttr(err))
		}
	}()

	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("decompress function failed to read from the writter: %w", err)
	}

	return b.Bytes(), nil
}

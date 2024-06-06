package compressor

import (
	"bytes"
	"compress/flate"
	"fmt"
	"log"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := flate.NewWriter(&b, flate.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return b.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer func() {
		if err := r.Close(); err != nil {
			log.Println("failed to close the reader")
		}
	}()

	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("decompress function failed to read from the writter: %w", err)
	}

	return b.Bytes(), nil
}

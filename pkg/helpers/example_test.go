package helpers

import (
	"context"
	"errors"
	"log/slog"
)

func ExampleErrAttr() {
	err := errors.New("example error")
	attr := ErrAttr(err)
	logger := slog.Default()
	logger.InfoContext(context.Background(), "example error", attr)

	// Output:
	// {"level":"INFO","msg":"example error","error":"example error"}
}

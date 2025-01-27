package helpers

import "log/slog"

// ErrAttr creates a structured logging attribute for errors using slog.
// It returns a slog.Attr that can be used in logging statements.
//
// If the error is nil, it returns an empty slog.Attr.
// If the error is not nil, it returns a slog.Attr with key "error" and the error as value.
//
// Example usage:
//
//	logger.Info("operation failed", helpers.ErrAttr(err))
func ErrAttr(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.Any("error", err)
}

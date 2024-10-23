package helpers

import "log/slog"

func ErrAttr(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.Any("error", err)
}

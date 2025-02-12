package utils

// Replace replaces the value of a pointer with a new value.
func Replace[T any](s *T, r T) {
	*s = r
}

package helpers

// Int64Ptr function to create pointers to int64 and float64 values.
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the provided float64 value.
func Float64Ptr(f float64) *float64 {
	return &f
}

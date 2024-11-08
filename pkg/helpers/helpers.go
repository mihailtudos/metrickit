package helpers

// Helper function to create pointers to int64 and float64 values.
func Int64Ptr(i int64) *int64 {
	return &i
}

func Float64Ptr(f float64) *float64 {
	return &f
}

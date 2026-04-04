package convert

func SafeDeref[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}

	return *ptr
}

func SafePtr[T any](val T) *T {
	return &val
}

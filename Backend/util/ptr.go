package util

func Ptr[T any](p T) *T {
	return &p
}

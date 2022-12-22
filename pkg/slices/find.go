package slices

func Find[T comparable](value T, slice []T) int {
	for i := 0; i < len(slice); i++ {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

package utils

func IsValueInList[T comparable](needle T, in []T) bool {
	for _, f := range in {
		if needle == f {
			return true
		}
	}
	return false
}

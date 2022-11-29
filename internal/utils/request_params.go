package utils

func GetValueByPointerOrDefaultValue[T any](param *T, defaultValue T) T {
	if param != nil {
		return *param
	}

	return defaultValue
}

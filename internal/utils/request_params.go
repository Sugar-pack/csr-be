package utils

func GetParamInt(param *int64, defaultValue int) int {
	if param != nil {
		return int(*param)
	}
	return defaultValue
}

func GetParamString(param *string, defaultValue string) string {
	if param != nil {
		return *param
	}
	return defaultValue
}

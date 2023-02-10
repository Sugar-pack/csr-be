package operations

func (a *BeAPI) GetExistingEndpoints() map[string][]string {
	result := make(map[string][]string)
	for method, pathToHandler := range a.handlers {
		for path := range pathToHandler {
			result[method] = append(result[method], path)
		}
	}
	return result
}

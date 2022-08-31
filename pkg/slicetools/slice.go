package slicetools

func ToInterface[T any](slice []T) []interface{} {
	iSlice := make([]interface{}, len(slice))
	for i, element := range slice {
		iSlice[i] = element
	}

	return iSlice
}

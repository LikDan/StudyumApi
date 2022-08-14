package utils

import "reflect"

func ToInterfaceSlice[T any](slice []T) []interface{} {
	var interface_ []interface{}
	for _, element := range slice {
		interface_ = append(interface_, element)
	}

	return interface_
}

func SliceContains[T any](slice []T, element T) bool {
	for _, t := range slice {
		if reflect.DeepEqual(element, t) {
			return true
		}
	}
	return false
}

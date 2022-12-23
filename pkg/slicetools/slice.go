package slicetools

import "golang.org/x/exp/slices"

func ToInterface[T any](slice []T) []interface{} {
	iSlice := make([]interface{}, len(slice))
	for i, element := range slice {
		iSlice[i] = element
	}

	return iSlice
}

func RemoveSameFunc[E any](slice1 []E, slice2 []E, compare func(E, E) bool) ([]E, []E) {
	offset1 := 0
	offset2 := 0

	for i1 := range slice1 {
		for i2 := range slice2 {
			if i1 < offset1 || i2 < offset2 {
				continue
			}

			if compare(slice1[i1-offset1], slice2[i2-offset2]) {
				slice1 = slices.Delete(slice1, i1-offset1, i1-offset1+1)
				slice2 = slices.Delete(slice2, i2-offset2, i2-offset2+1)

				offset1 += 1
				offset2 += 1
			}
		}
	}

	return slice1, slice2
}

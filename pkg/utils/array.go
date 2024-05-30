package utils

// Function to remove multiple elements by indices
func RemoveByIndex[T any](source []T, index []int) []T {
	var tmpArr []T

	for i := range source {
		var foundIndex = false
		for _, rIndex := range index {
			if i == rIndex {
				foundIndex = true
				break
			}
		}

		if foundIndex {
			continue
		}

		tmpArr = append(tmpArr, source[i])
	}

	return tmpArr
}

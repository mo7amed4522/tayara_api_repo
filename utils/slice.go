package utils

// Intersect Find the intersection of two integer slices.
func Intersect(slice1, slice2 []int) []int {
	intersectMap := make(map[int]bool)
	var intersection []int

	// Traverse the first slice, record each element
	for _, value := range slice1 {
		intersectMap[value] = true
	}

	// Traverse the second slice, check if the element is in the map
	for _, value := range slice2 {
		if _, exists := intersectMap[value]; exists {
			intersection = append(intersection, value)
			// Delete elements from the map to avoid duplicate addition
			delete(intersectMap, value)
		}
	}

	return intersection
}

// Intersect Find the difference between two int type slices
func Difference(slice1, slice2 []int) []int {
	diff := []int{}
	map2 := make(map[int]bool)
	for _, v := range slice2 {
		map2[v] = true
	}
	for _, v := range slice1 {
		if !map2[v] {
			diff = append(diff, v)
		}
	}
	return diff
}

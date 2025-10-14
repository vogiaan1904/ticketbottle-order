package util

func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// Intersection returns elements that are in both slice a and slice b
func Intersection(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var intersection []string
	for _, x := range a {
		if _, found := mb[x]; found {
			intersection = append(intersection, x)
		}
	}
	return intersection
}

// EqualSlices checks if two string slices are equal (same elements in same order)
func EqualSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

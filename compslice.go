package scheduling

// ComparableSlice : would be used to keep relay ids
// we would want to know if there are common ids and some context we would need if all are matching
type ComparableSlice []string

// Intersection : tries to get the intersecting items from 2 slices
// items that are common to both
// items that are unique to cmpsl
// items that are unique to other
func (cmpsl ComparableSlice) Intersection(other ComparableSlice) (int, int, int) {
	comm := 0
	for _, item := range cmpsl {
		for _, oitem := range other {
			if item == oitem {
				comm++
			}
		}
	}
	return comm, len(cmpsl) - comm, len(other) - comm
}

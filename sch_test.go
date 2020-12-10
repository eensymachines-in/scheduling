package scheduling

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparableSlice(t *testing.T) {
	sl1 := ComparableSlice{"IN1", "IN2", "IN3"}
	sl2 := ComparableSlice{"IN2", "IN3", "IN4", "IN1"}
	matches, mismatch1, mismatch2 := sl1.Intersection(sl2)
	assert.Equal(t, 3, matches, "Was expecting 2 matches in the slices above")
	assert.Equal(t, 0, mismatch1, "Incorrect mismatches on the first")
	assert.Equal(t, 1, mismatch2, "Incorrect mismatches on the second")

	t.Log("--------------------------\n")
	sl1 = ComparableSlice{}
	sl2 = ComparableSlice{"IN2", "IN3", "IN4", "IN1"}
	matches, mismatch1, mismatch2 = sl1.Intersection(sl2)
	assert.Equal(t, 0, matches, "Was expecting 2 matches in the slices above")
	assert.Equal(t, 0, mismatch1, "Incorrect mismatches on the first")
	assert.Equal(t, 4, mismatch2, "Incorrect mismatches on the second")
}

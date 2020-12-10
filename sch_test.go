package scheduling

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

func TestScheduleRead(t *testing.T) {
	jsonFile, err := os.Open("test_sched.json")
	if err != nil {
		t.Error(err)
		return
	}
	// Reading bytes from the file and unmarshalling the same to struct values
	bytes, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
	type conf struct {
		Schedules []JSONRelayState `json:"schedules"`
	}
	c := conf{}
	if json.Unmarshal(bytes, &c) != nil {
		t.Error("Failed to unmarshal json from file")
		return
	}
	t.Log("Now logging the JSONRelayState")
	t.Log(c)
	for _, s := range c.Schedules {
		sched, err := s.ToSchedule()
		if err != nil {
			t.Errorf("Error coverting to schedule objecy %s", err)
			return
		}
		t.Log(sched)
	}
}

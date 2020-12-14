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
	for _, s := range c.Schedules {
		sched, err := s.ToSchedule()
		if err != nil {
			t.Errorf("Error coverting to schedule objecy %s", err)
			return
		}
		t.Log(sched)
	}
}

// func TestScheduleNrFrTrigger(t *testing.T) {
// 	jsonFile, _ := os.Open("test_sched.json")
// 	// Reading bytes from the file and unmarshalling the same to struct values
// 	bytes, _ := ioutil.ReadAll(jsonFile)
// 	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
// 	type conf struct {
// 		Schedules []JSONRelayState `json:"schedules"`
// 	}
// 	c := conf{}
// 	json.Unmarshal(bytes, &c)
// 	for _, s := range c.Schedules {
// 		sched, err := s.ToSchedule()
// 		if err != nil {
// 			t.Errorf("Error coverting to schedule objecy %s", err)
// 			return
// 		}
// 		t.Logf("Schedule: %s", sched)
// 		nr, fr, pre, post := sched.NearFarTrigger(elapsedSecondsNow())
// 		t.Logf("Near trigger: %s", nr)
// 		t.Logf("Far trigger: %s", fr)
// 		t.Logf("Pre sleep: %d", pre)
// 		t.Logf("Post sleep: %d", post)
// 	}
// }

/*This test can let you know if primary and secondary schedules correctly report */
func TestScheduleConflicts(t *testing.T) {
	jsonFile, _ := os.Open("test_sched2.json")
	// Reading bytes from the file and unmarshalling the same to struct values
	bytes, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
	type conf struct {
		Schedules []JSONRelayState `json:"schedules"`
	}
	c := conf{}
	json.Unmarshal(bytes, &c)
	primaSched, err := c.Schedules[0].ToSchedule()
	if err != nil {
		t.Error(err)
		panic("Failed to read the primary schedule")
	}
	for i, s := range c.Schedules {
		if i > 0 {
			// since we want to compare all with primary schedule
			sched, err := s.ToSchedule()
			if err != nil {
				t.Errorf("Error coverting to schedule object %s", err)
				return
			}
			t.Logf("Schedule: %s", sched)
			t.Logf("Conflicts: %t", primaSched.ConflictsWith(sched))
			t.Logf("Delay: %d", sched.Delay())
		}

	}
}

/*Here we test conflicts of 2 or more patch schedules amongst each other*/
func TestSchedulePatchConflicts(t *testing.T) {
	jsonFile, _ := os.Open("test_sched2.json")
	// Reading bytes from the file and unmarshalling the same to struct values
	bytes, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
	type conf struct {
		Schedules []JSONRelayState `json:"schedules"`
	}
	c := conf{}
	json.Unmarshal(bytes, &c)
	// We are leaving out the primary schedule
	// would test patch schedules wrt to other patch schedules
	patchSched, err := c.Schedules[2].ToSchedule()
	if err != nil {
		t.Error(err)
		panic("Failed to read the first patch schedule ")
	}
	for i, s := range c.Schedules {
		if i > 2 {
			// since we want to compare all with primary schedule
			sched, err := s.ToSchedule()
			if err != nil {
				t.Errorf("Error coverting to schedule object %s", err)
				return
			}
			t.Logf("Schedule: %s", sched)
			t.Logf("Conflicts: %t", patchSched.ConflictsWith(sched))
			t.Logf("Delay: %d", sched.Delay())
		}

	}
}

func TestScheduleApply(t *testing.T) {
	jsonFile, _ := os.Open("test_sched3.json")
	// Reading bytes from the file and unmarshalling the same to struct values
	bytes, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close() // since this returns a closure, the call to this cannot be deferred
	type conf struct {
		Schedules []JSONRelayState `json:"schedules"`
	}
	c := conf{}
	json.Unmarshal(bytes, &c)
	scheds := []Schedule{}
	for _, s := range c.Schedules {
		sched, err := s.ToSchedule()
		if err != nil {
			t.Errorf("Error coverting to schedule object %s", err)
			return
		}
		scheds = append(scheds, sched)
	}
	for i, s := range scheds {
		for _, ss := range scheds[i+1:] {
			if s.ConflictsWith(ss) {
				ss.AddConflict()
			}
		}
	}
	ok, cancel := make(chan interface{}, 10), make(chan interface{})
	send := make(chan []byte, 10)
	err := make(chan error, 10)
	defer close(ok)
	defer close(cancel)
	defer close(send)
	defer close(err)
	go func(listen chan []byte) {
		for msg := range listen {
			t.Log(string(msg))
		}
	}(send)
	for _, s := range scheds {
		if s.Conflicts() == 0 {
			task := s.ToTask(elapsedSecondsNow())
			go task.Apply(ok, cancel, send, err)
		}
	}
	for okmsg := range ok {
		t.Logf("end of a schedule.. %v", okmsg)
	}
}

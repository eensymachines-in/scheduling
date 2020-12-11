package scheduling

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTrigger : lets test simple trigger functions
func TestTrigger(t *testing.T) {

	t.Log("---------------------------\n")
	t.Log("Now testing trigger relayids ")
	rs1 := &RelayState{byte(0), "IN1"}
	rs2 := &RelayState{byte(0), "IN2"}
	trigg := NewTrg(64800, rs1, rs2)
	t.Logf("Relay IDs for trigger %v", trigg.RelayIDs())
	t.Logf("Number of relays controlled with trigger %d", trigg.RelayCount())
	byt, _ := json.Marshal(trigg)
	t.Logf("Trigger message over TCP %s", string(byt))
	assert.Equal(t, true, trigg.HasRelayWithID("IN1"), "Was expecting IN1 to be inside the trigger")
	assert.Equal(t, true, trigg.HasRelayWithID("IN2"), "Was expecting IN2 to be inside the trigger")
	assert.Equal(t, false, trigg.HasRelayWithID("IN3"), "Was not expecting IN3 to be inside the trigger")

	t.Log("---------------------------\n")
	t.Log("Now testing Trigger creation with duplicate relay ids ")
	// this third one shoudl not be considered since relay ids in triggers are unique
	rs3 := &RelayState{byte(1), "IN1"}
	trigg = NewTrg(64800, rs1, rs2, rs3)
	byt, _ = json.Marshal(trigg)
	t.Logf("Trigger message over TCP %s", string(byt))
	t.Logf("Relay IDs for trigger %v", trigg.RelayIDs())
	t.Logf("Number of relays controlled with trigger %d", trigg.RelayCount())
	t.Logf("Trigger message over TCP %s", string(byt))
	assert.Equal(t, true, trigg.HasRelayWithID("IN1"), "Was expecting IN1 to be inside the trigger")
	assert.Equal(t, true, trigg.HasRelayWithID("IN2"), "Was expecting IN2 to be inside the trigger")
	assert.Equal(t, false, trigg.HasRelayWithID("IN3"), "Was not expecting IN3 to be inside the trigger")

	t.Log("---------------------------\n")
	t.Log("Now for intersection test for the 2 triggers")
	rs1 = &RelayState{byte(0), "IN1"}
	rs2 = &RelayState{byte(0), "IN2"}
	rs3 = &RelayState{byte(0), "IN2"}
	rs4 := &RelayState{byte(0), "IN3"}
	trigg1 := NewTrg(64800, rs1, rs2)
	trigg2 := NewTrg(64800, rs3, rs4)
	trigg3 := NewTrg(64800, rs1, rs2) //exactly intersecting with trigg1
	assert.Equal(t, true, trigg1.Intersects(trigg2, false), "Was expecting the triggers to be intersecting")
	assert.Equal(t, false, trigg1.Intersects(trigg2, true), "Was not expecting the triggers to intersection in the exact mode")
	assert.Equal(t, true, trigg1.Intersects(trigg3, true), "Was expecting trigg1 and trig3 to be intersecting in the exact mode")
}

package scheduling

import (
	"encoding/json"
	"fmt"
)

// rlyStateTrg : trigger is just a timestamp and collection of relay state
type rlyStateTrg struct {
	at int
	rs []*RelayState
}

// Trigger : interface usd by schedule to talk to trigger objects
type Trigger interface {
	// SendTCP(serverip, port string) error
	At() int
	FlipAllRelays()
	RelayCount() int
	RelayIDs() ComparableSlice
	HasRelayWithID(find string) bool
	// Helps to see if 2 triggers are operating on the same or common relays
	// if strict : length of the array also comes in play and checks to see if the mismatches are equal too
	Intersects(other Trigger, exact bool) bool
	// Checks to see if the trigger is coincident on time
	Coincides(other Trigger) bool
}

func (tr *rlyStateTrg) String() string {
	return fmt.Sprintf("%s, %v", TmStrFromUnixSecs(tr.at), tr.RelayIDs())
}
func (tr *rlyStateTrg) RelayIDs() ComparableSlice {
	result := ComparableSlice{}
	for _, item := range tr.rs {
		result = append(result, item.ID())
	}
	return result
}

// Intersects : finds if there are relays in 2 triggers that have the same id
// in exact mode, all the relay ids shoudl be matching
func (tr *rlyStateTrg) Intersects(other Trigger, exact bool) bool {
	if other == nil {
		// If there's nothing to intersct there isnt any intersection at all
		return false
	}
	// We have to get both the relay states as comparable slices
	cmpSli1, cmpSli2 := tr.RelayIDs(), other.RelayIDs()
	matches, mm1, mm2 := cmpSli1.Intersection(cmpSli2)
	if matches > 0 {
		if exact {
			if mm1 == 0 && mm2 == 0 {
				// All the relays are exactly matching , no mismatches founds
				return true
			}
			return false
		}
		return true
	}
	return false
}

func (tr *rlyStateTrg) Coincides(other Trigger) bool {
	if tr.At() == other.At() {
		return true
	}
	return false
}

// HasRelayWithID : scans thru all the relay states to know if the ID of the relay exists
func (tr *rlyStateTrg) HasRelayWithID(find string) bool {
	for _, state := range tr.rs {
		if state.ID() == find {
			return true
		}
	}
	return false
}

// MarshalJSON : overriding the default implementation of marshaling json
// this can help us send thru TCP with much ease
func (tr *rlyStateTrg) MarshalJSON() ([]byte, error) {
	mpResult := map[string]byte{}
	for _, state := range tr.rs {
		for k, v := range state.Status() {
			mpResult[k] = v
		}
	}
	return json.Marshal(mpResult)
}

// FlipAllRelays : Flips all relays contained within, composite function sugar coat
func (tr *rlyStateTrg) FlipAllRelays() {
	for _, state := range tr.rs {
		state.Flip()
	}
}

// // SendTCP : wires the trigger over TCP to sockets
// func (tr *rlyStateTrg) SendTCP(serverip, port string) error {
// 	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", serverip, port))
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	data, _ := json.Marshal(tr)
// 	_, err = conn.Write(data)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// At : seconds since midnight at which the trigger becomes effective
func (tr *rlyStateTrg) At() int {
	return tr.at
}
func (tr *rlyStateTrg) RelayCount() int {
	return len(tr.rs)
}

// NewTrg : makes a new trigger  with variadic number of relays
// A single trigger can have unique relay ids only
func NewTrg(secs int, states ...*RelayState) Trigger {
	result := rlyStateTrg{secs, []*RelayState{}}
	for _, s := range states {
		if !result.HasRelayWithID(s.ID()) {
			// Cause relay states with same ID cannot be added to the same trigger
			// ID of the relay state inside a trigger is unique
			result.rs = append(result.rs, s)
		}
	}
	return &result
}

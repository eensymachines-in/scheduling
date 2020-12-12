package scheduling

// RelayState : this is just to hold the state of relay with the identification of the relay
// relay should be identified with the same name as required by srvrelay
// Storing this as just a byte is also possible, but that is when we want the relay module to work as a block, not when we want to operate on individual relays
type RelayState struct {
	state byte
	id    string
}

// Status : Gets the state of the relay with ID
func (rs *RelayState) Status() map[string]byte {
	return map[string]byte{rs.id: rs.state}
}

// Flip : flips the state of the relay
func (rs *RelayState) Flip() {
	rs.state = byte(1) - rs.state
}

// State : sets the state of the relay
func (rs *RelayState) State(new byte) *RelayState {
	if new > 0 {
		rs.state = byte(1)
	}
	rs.state = byte(0)
	return rs
}

// ID : spits out the id of the relay state
// this is generally the relay ID on the actual relay, IN1, IN2, IN3..
func (rs *RelayState) ID() string {
	return rs.id
}

// NewRelayState : quick way to make a new relay state
func NewRelayState(id string) *RelayState {
	return &RelayState{byte(0), id}
}

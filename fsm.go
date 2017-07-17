package ssfsm

import (
	"github.com/pkg/errors"
)

type serr string

func (v serr) Error() string { return string(v) }

// Main Cause Errors
var (
	ErrTransitionConflict error = serr("transition_conflict: another transition is in progress")
	ErrEventNotFound      error = serr("event_not_found")
	ErrStateConflict      error = serr("state_conflict: this event triggers no state")
)

// Transition represents a transition
type Transition struct {
	Event, From, To string
}

// FSM is a finite state machine
type FSM struct {
	mx    *chan struct{}
	state string
	graph map[string]map[string]next
}

type next struct {
	transition Transition
	callback   func(*FSM, Transition) error
}

// Trigger triggers an event. If async is set to true, then Trigger would block.
// And if any other transitions are in progress, it will return an error. This implies
// that Trigger must not get called recursively.
func (sm *FSM) Trigger(event string) error {
	if sm.mx != nil {
		select {
		case *sm.mx <- struct{}{}:
			defer func() { <-*sm.mx }()
		default:
			return ErrTransitionConflict
		}
	}
	vx, ok := sm.graph[event]
	if !ok {
		return errors.Wrap(ErrEventNotFound, "event: "+event)
	}
	nx, ok := vx[sm.state]
	if !ok {
		return errors.Wrap(ErrStateConflict, "event: "+event+" current state: "+sm.state)
	}
	var err error
	starting := sm.state
	defer func() {
		if err != nil {
			return
		}
		if starting != sm.state {
			return
		}
		sm.state = nx.transition.To
	}()
	if nx.callback == nil {
		return nil
	}
	err = nx.callback(sm, nx.transition)
	return err
}

// Table represents a transition table between states
type Table map[Transition]func(*FSM, Transition) error

// NewFSM creates an instance of FSM. If async is set to true, then Trigger would block.
// And if any other transitions are in progress, it will return an error.
func NewFSM(async bool, state string, table Table) *FSM {
	// event -> from -> (to, callback)
	var graph = make(map[string]map[string]next)

	for gk, gv := range table {
		vx, ok := graph[gk.Event]
		if !ok {
			vx = make(map[string]next)
		}
		vx[gk.From] = next{gk, gv}
		graph[gk.Event] = vx
	}

	res := &FSM{
		state: state,
		graph: graph,
	}
	if async {
		cb := make(chan struct{}, 1)
		res.mx = &cb
	}

	return res
}

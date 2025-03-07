package gomata

import (
	"errors"
)

type State struct {
	Value string
}
type Event struct {
	Type  string
	State string
}
type StateNode struct {
	ID      string               `json:"id"`
	Initial string               `json:"initial"`
	States  map[string]StateNode `json:"states"`
	Entry   string               `json:"entry"`
	Exit    string               `json:"exit"`
	On      map[string]string    `json:"on"`

	subscribers      []*func(Event)
	currentState     *StateNode
	currentStateName string
	emit             func(Event)

	// TODO: Should this be consolidated with subscribers and use internal flags to codnitionally fire emits and transitions?
	transitionSubscribers []*func(State)
	stateChange           func(State)
}

func (s *StateNode) Init() error {
	// check if initial state set
	if s.Initial != "" {
		// check if initial state exists in 'states'
		state, ok := s.States[s.Initial]
		// if it does not; there's miss-config
		if !ok {
			return errors.New("No definition for state: " + s.Initial + ". Remove initial state or add it to 'states' config.")
		}
		// define event handler and subscribe to the state node
		// cannot pass func method pointer; needs to be wrapped
		s.emit = func(e Event) {
			s.Emit(e)
		}
		s.stateChange = func(state State) {
			s.EmitTransition()
		}
		state.Subscribe(&s.emit)
		state.SubscribeToTransitions(&s.stateChange)
		s.setCurrentState(s.Initial, &state)
	}
	// emit enter transition if defined
	s.enter()

	// if current state was set, initialize it
	if s.currentState != nil {
		return s.currentState.Init()
	}

	return nil
}
func (s *StateNode) GetState() string {
	if s.currentState != nil {
		childState := s.currentState.GetState()
		if childState != "" {
			return s.currentStateName + "." + childState
		}
		return s.currentStateName
	}
	return ""
}
func (s *StateNode) setCurrentState(name string, node *StateNode) {
	if name == "" || node == nil {
		return
	}
	s.currentStateName = name
	s.currentState = node
	s.EmitTransition()
}
func (s *StateNode) Close() {
	// emit exit if exists
	s.exit()
	// dump all subscribers
	s.subscribers = []*func(Event){}
}
func (s *StateNode) transition(event string) bool {
	// eventCaptured flag is used to capture the event
	// state node will first check if current state node can handle the event
	// if it can, child handles event and parent does not.
	var eventCaptured bool
	if s.currentState != nil {
		eventCaptured = s.currentState.transition(event)
	}

	// if event was not capture by current state node try to handle it
	if !eventCaptured {
		// if no current state node set, theres nothing to handle
		if s.currentState == nil {
			return false
		}
		// if theres no event for the transition - nothing to handle
		nextState, ok := s.currentState.On[event]
		if !ok {
			return false
		}

		// if nextState does not map to a state node - nothign to handle
		nextStateNode, exists := s.States[nextState]
		if !exists {
			return false
		}
		// cleanup
		s.currentState.Close()

		// Init current state node
		nextStateNode.Subscribe(&s.emit)
		nextStateNode.SubscribeToTransitions(&s.stateChange)
		s.setCurrentState(nextState, &nextStateNode)
		s.currentState.Init()
		return true
	}
	return true
}
func (s *StateNode) enter() {
	// emit entry event if exists
	if s.Entry != "" {
		s.Emit(Event{Type: s.Entry})
	}
}
func (s *StateNode) exit() {
	// emit exit if exists; first see if there's a current state node that needs to handle exit
	if s.currentState != nil {
		s.currentState.exit()
	}
	if s.Exit != "" {
		s.Emit(Event{Type: s.Exit})
	}
}
func (s *StateNode) Subscribe(cb *func(Event)) {
	// subscribe to state node's events
	s.subscribers = append(s.subscribers, cb)
}
func (s *StateNode) Unsubscribe(cb *func(Event)) {
	// unsubscribe using specific fun pointer
	newSubs := []*func(Event){}
	for _, fn := range s.subscribers {
		if fn != cb {
			newSubs = append(newSubs, fn)
		}
	}
	s.subscribers = newSubs

}
func (s *StateNode) Emit(e Event) {
	// emit event to all the subscribers
	if len(s.subscribers) != 0 {
		for _, fn := range s.subscribers {
			(*fn)(e)
		}
	}

}
func (s *StateNode) SubscribeToTransitions(cb *func(e State)) {
	s.transitionSubscribers = append(s.transitionSubscribers, cb)
}
func (s *StateNode) UnsubscribeFromTransitions(cb *func(State)) {
	// unsubscribe using specific fun pointer
	newSubs := []*func(State){}
	for _, fn := range s.transitionSubscribers {
		if fn != cb {
			newSubs = append(newSubs, fn)
		}
	}
	s.transitionSubscribers = newSubs

}
func (s *StateNode) EmitTransition() {
	if len(s.transitionSubscribers) != 0 {
		for _, fn := range s.transitionSubscribers {
			(*fn)(State{Value: s.GetState()})
		}
	}
}

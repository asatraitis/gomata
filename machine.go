package gomata

import (
	"encoding/json"
	"sync"
)

type Machine struct {
	mutx   sync.Mutex
	root   *StateNode
	config string
}

func NewMachine(jsonConfig string) *Machine {
	return &Machine{
		mutx:   sync.Mutex{},
		root:   &StateNode{},
		config: jsonConfig,
	}
}
func (m *Machine) Start() error {
	err := json.Unmarshal([]byte(m.config), m.root)
	if err != nil {
		return err
	}

	m.mutx.Lock()
	defer m.mutx.Unlock()
	return m.root.Init()
}
func (m *Machine) OnEmit(handler *func(Event)) {
	m.root.Subscribe(handler)
}
func (m *Machine) OnTransition(handler *func(State)) {
	m.root.SubscribeToTransitions(handler)
}
func (m *Machine) GetState() string {
	m.mutx.Lock()
	defer m.mutx.Unlock()
	return m.root.GetState()
}
func (m *Machine) Send(event string) {
	m.mutx.Lock()
	defer m.mutx.Unlock()
	m.root.transition(event)
}

package gomata

import (
	"encoding/json"
	"sync"
)

type Machine struct {
	mutx   sync.Mutex
	root   *StateNode
	config string

	handler *func(Event)
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
func (m *Machine) OnTransition(handler *func(Event)) {
	m.handler = handler
	m.root.Subscribe(m.handler)
}
func (m *Machine) Send(event string) {

	m.mutx.Lock()
	defer m.mutx.Unlock()
	m.root.transition(event)
}

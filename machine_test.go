package gomata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMachine(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle",
		"states": {
			"idle": {}
		}
	}`
	m := NewMachine(jsonConfig)
	assert.Equal(t, jsonConfig, m.config)
	assert.NotNil(t, m.root)
}

func TestStart(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle",
		"states": {
			"idle": {}
		}
	}`
	m := NewMachine(jsonConfig)
	err := m.Start()
	assert.NoError(t, err)
	assert.Equal(t, "idle", m.root.currentStateName)
}

func TestStart_FAIL_BadJSON(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle"
		"states": {
			"idle": {}
		}
	}`
	m := NewMachine(jsonConfig)
	err := m.Start()
	assert.Error(t, err)
}

func TestStart_FAIL_BadInitialState(t *testing.T) {
	const jsonConfig = `{
		"initial": "pending",
		"states": {
			"idle": {}
		}
	}`
	m := NewMachine(jsonConfig)
	err := m.Start()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "No definition for state: pending")
}

func TestOnEmit(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle",
		"states": {
			"idle": {}
		}
	}`
	m := NewMachine(jsonConfig)
	handler := func(e Event) {}
	assert.Len(t, m.root.subscribers, 0)
	m.OnEmit(&handler)
	assert.Len(t, m.root.subscribers, 1)
}

func TestSend(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle",
		"states": {
			"idle": {
				"entry": "entered_idle",
				"on": {
					"START": "running"
				}
			},
			"running": {
				"entry": "entered_running"
			}
		}
	}`
	m := NewMachine(jsonConfig)
	testEvents := []Event{}
	handler := func(e Event) {
		testEvents = append(testEvents, e)
	}
	m.OnEmit(&handler)
	err := m.Start()
	assert.NoError(t, err)
	assert.Len(t, testEvents, 1)
	assert.Equal(t, "entered_idle", testEvents[0].Type)
	assert.Equal(t, "idle", m.GetState())
	m.Send("START")
	assert.Len(t, testEvents, 2)
	assert.Equal(t, "running", m.GetState())
	assert.Equal(t, "entered_running", testEvents[1].Type)
}

func TestOnTransition_OK(t *testing.T) {
	const jsonConfig = `{
		"initial": "idle",
		"states": {
			"idle": {
				"on": {
					"START": "running"
				}
			},
			"running": {}
		}
	}`
	m := NewMachine(jsonConfig)
	testStates := []State{}
	handler := func(state State) {
		testStates = append(testStates, state)
	}

	m.OnTransition(&handler)
	m.Start()
	assert.Len(t, testStates, 1)
	assert.Equal(t, "idle", testStates[0].Value)

	m.Send("START")
	assert.Len(t, testStates, 2)
	assert.Equal(t, "running", testStates[1].Value)
}

package gomata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeInit_OK(t *testing.T) {
	// empty node
	node := StateNode{}
	err := node.Init()
	assert.NoError(t, nil, err)

	// node with single state
	node.Initial = "idle"
	node.States = map[string]StateNode{"idle": StateNode{}}
	err = node.Init()
	assert.NoError(t, nil, err)
	assert.Equal(t, "idle", node.currentStateName)

	// nested
	node.Initial = "idle"
	node.States = map[string]StateNode{
		"idle": StateNode{
			ID:      "idle-node",
			Initial: "low",
			States: map[string]StateNode{
				"low": StateNode{
					ID: "low-node",
				},
			},
		},
	}
	err = node.Init()
	assert.NoError(t, nil, err)
	assert.Equal(t, "idle", node.currentStateName)
	assert.Equal(t, "idle-node", node.currentState.ID)
	assert.Equal(t, "low", node.currentState.currentStateName)
	assert.Equal(t, "low-node", node.currentState.currentState.ID)
}

func TestNodeInit_FAIL(t *testing.T) {
	// Initial does not match state
	node := StateNode{}
	node.Initial = "idling"
	node.States = map[string]StateNode{"idle": StateNode{}}
	err := node.Init()
	assert.Error(t, err)
	assert.Equal(t, "", node.currentStateName)

	// state def does not match initial state
	node.Initial = "idle"
	node.States = map[string]StateNode{"idling": StateNode{}}
	err = node.Init()
	assert.Error(t, err)
	assert.Equal(t, "", node.currentStateName)
}

func TestSetCurrentState_OK(t *testing.T) {
	// not nil name and node
	node := StateNode{}
	node.setCurrentState("whatever", &StateNode{ID: "whatever-node"})
	assert.Equal(t, "whatever-node", node.currentState.ID)
	assert.Equal(t, "whatever", node.currentStateName)

	// nil name and not-nil node: sets when both are valid
	node = StateNode{}
	node.setCurrentState("", &StateNode{})
	assert.Nil(t, node.currentState)
	assert.Empty(t, node.currentStateName)

	// not-nil name and nil node: sets when both are valid
	node = StateNode{}
	node.setCurrentState("test", nil)
	assert.Nil(t, node.currentState)
	assert.Empty(t, node.currentStateName)
}

func TestClose_OK(t *testing.T) {
	var exitEvent Event
	handler := func(e Event) {
		exitEvent = e
	}
	node := StateNode{
		Exit:        "exit_test_event",
		subscribers: []*func(Event){&handler},
	}
	assert.Len(t, node.subscribers, 1)
	node.Close()
	assert.Len(t, node.subscribers, 0)
	assert.Equal(t, "exit_test_event", exitEvent.Type)
}

func TestEnter_OK(t *testing.T) {
	var enterEvent Event
	handler := func(e Event) {
		enterEvent = e
	}
	node := StateNode{
		Entry:       "enter_test_event",
		subscribers: []*func(Event){&handler},
	}
	node.enter()
	assert.Equal(t, "enter_test_event", enterEvent.Type)
}

func TestExit_OK(t *testing.T) {
	var testEvent Event
	handler := func(e Event) {
		testEvent = e
	}
	node := StateNode{
		Exit:        "exit_test",
		subscribers: []*func(Event){&handler},
	}
	node.exit()
	assert.Equal(t, "exit_test", testEvent.Type)

	testEvents := []Event{}
	handler = func(e Event) {
		testEvents = append(testEvents, e)
	}

	node = StateNode{
		Exit:    "exit_parent",
		Initial: "idle",
		States: map[string]StateNode{
			"idle": StateNode{
				Exit: "exit_child",
			},
		},
		subscribers: []*func(Event){&handler},
	}
	// Init sets-up handlers between states
	node.Init()
	node.exit()
	assert.Len(t, testEvents, 2)
	// exit events start at children and bubble up to parent
	assert.Equal(t, "exit_child", testEvents[0].Type)
	assert.Equal(t, "exit_parent", testEvents[1].Type)
}

func TestSubscribe_OK(t *testing.T) {
	node := StateNode{}
	handler := func(e Event) {}
	handler2 := func(e Event) {}

	node.Subscribe(&handler)
	assert.Len(t, node.subscribers, 1)

	node.Subscribe(&handler2)
	assert.Len(t, node.subscribers, 2)
}

func TestUnsubscribe_OK(t *testing.T) {
	node := StateNode{}
	handler := func(e Event) {}
	handler2 := func(e Event) {}

	node.Subscribe(&handler)
	node.Subscribe(&handler2)
	assert.Len(t, node.subscribers, 2)

	node.Unsubscribe(&handler)
	assert.Len(t, node.subscribers, 1)

	node.Unsubscribe(&handler2)
	assert.Len(t, node.subscribers, 0)
}

func TestEmit_OK(t *testing.T) {
	node := StateNode{}
	var testEvent Event
	handler := func(e Event) {
		testEvent = e
	}
	node.Subscribe(&handler)
	node.Emit(Event{Type: "test_event"})
	assert.Equal(t, "test_event", testEvent.Type)
}

func TestTransition(t *testing.T) {
	testEvents := []Event{}
	handler := func(e Event) {
		testEvents = append(testEvents, e)
	}
	// flat structure; transitions at top level
	node := StateNode{
		Initial: "idle",
		States: map[string]StateNode{
			"idle": StateNode{
				Entry: "entered_idle",
				On: map[string]string{
					"START": "running",
				},
			},
			"running": StateNode{
				Entry: "entered_running",
				On: map[string]string{
					"STOP": "idle",
				},
			},
		},
	}
	node.Subscribe(&handler)
	node.Init()
	assert.Len(t, testEvents, 1)
	assert.Equal(t, "entered_idle", testEvents[0].Type)
	assert.Equal(t, "idle", node.currentStateName)

	node.transition("START")
	assert.Len(t, testEvents, 2)
	assert.Equal(t, "entered_running", testEvents[1].Type)
	assert.Equal(t, "running", node.currentStateName)

	node.transition("STOP")
	assert.Len(t, testEvents, 3)
	assert.Equal(t, "entered_idle", testEvents[2].Type)
	assert.Equal(t, "idle", node.currentStateName)

	// clear events
	// nested state nodes; transitions in child nodes
	testEvents = []Event{}
	node = StateNode{
		Initial: "idle",
		States: map[string]StateNode{
			"idle": StateNode{
				Initial: "idle.low",
				States: map[string]StateNode{
					"idle.low": StateNode{
						Entry: "entered_idle.low",
						On: map[string]string{
							"UP": "idle.high",
						},
					},
					"idle.high": StateNode{
						Entry: "entered_idle.high",
						On: map[string]string{
							"DOWN": "idle.low",
						},
					},
				},
			},
		},
	}
	node.Subscribe(&handler)
	node.Init()
	assert.Len(t, testEvents, 1)
	assert.Equal(t, "entered_idle.low", testEvents[0].Type)
	assert.Equal(t, "idle", node.currentStateName)
	assert.Equal(t, "idle.low", node.currentState.currentStateName)

	node.transition("UP")
	assert.Len(t, testEvents, 2)
	assert.Equal(t, "entered_idle.high", testEvents[1].Type)
	assert.Equal(t, "idle", node.currentStateName)
	assert.Equal(t, "idle.high", node.currentState.currentStateName)

	node.transition("DOWN")
	assert.Len(t, testEvents, 3)
	assert.Equal(t, "entered_idle.low", testEvents[2].Type)
	assert.Equal(t, "idle", node.currentStateName)
	assert.Equal(t, "idle.low", node.currentState.currentStateName)

	// transition event order
	// entering follows bubble down (parent event happens first then child)
	// exiting follows bubble up (child event happens first then parent)
	testEvents = []Event{}
	node = StateNode{
		Initial: "idle",
		States: map[string]StateNode{
			"idle": StateNode{
				Entry:   "entered_idle",
				Initial: "idle.low",
				On: map[string]string{
					"START": "running",
				},
				States: map[string]StateNode{
					"idle.low": StateNode{
						Entry: "entered_idle.low",
						Exit:  "exited_idle.low",
					},
				},
				Exit: "exited_idle",
			},
			"running": StateNode{
				Entry: "entered_running",
				On: map[string]string{
					"STOP": "idle",
				},
				Exit: "exited_running",
			},
		},
	}
	node.Subscribe(&handler)
	node.Init()
	// entering states follows bubble down (parent event happens first then child)
	assert.Len(t, testEvents, 2)
	assert.Equal(t, "entered_idle", testEvents[0].Type)
	assert.Equal(t, "entered_idle.low", testEvents[1].Type)
	// exiting states follows bubble up (child event happens first then parent)
	// "START" event will transition to "running" which first exit from idle.low then idle, then we will enter running in this order
	node.transition("START")
	assert.Len(t, testEvents, 5)
	assert.Equal(t, "exited_idle.low", testEvents[2].Type)
	assert.Equal(t, "exited_idle", testEvents[3].Type)
	assert.Equal(t, "entered_running", testEvents[4].Type)
}

func TestGetState_OK(t *testing.T) {
	node := StateNode{
		Initial: "idle",
		States: map[string]StateNode{
			"idle": StateNode{
				Entry:   "entered_idle",
				Initial: "idle.low",
				On: map[string]string{
					"START": "running",
				},
				States: map[string]StateNode{
					"idle.low": StateNode{
						Entry: "entered_idle.low",
						Exit:  "exited_idle.low",
					},
				},
				Exit: "exited_idle",
			},
			"running": StateNode{
				Entry: "entered_running",
				On: map[string]string{
					"STOP": "idle",
				},
				Exit: "exited_running",
			},
		},
	}
	node.Init()
	assert.Equal(t, "idle.idle.low", node.GetState())
}

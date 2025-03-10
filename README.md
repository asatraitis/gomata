# GoMata

[![License](https://img.shields.io/github/license/asatraitis/gomata)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/asatraitis/gomata)](https://goreportcard.com/report/github.com/asatraitis/gomata)

Lightweight and minimalistic Finite State Machine (FSM) implementation in Go. Allows parsing JSON string format similar to XState. Currently supports entry and exit events during transitions.

## Table of Contents

- [Installation](#installation)
- [Examples](#examples)

## Installation

```sh
go get github.com/asatraitis/gomata
```

## Examples

```go
package main

import (
	"fmt"

	"github.com/asatraitis/gomata"
)

func main() {
	const fsmJson = `{
		"initial": "standing",
		"states": {
			"standing": {
				"entry": "entered_standing", // event type for entry transition
				"on" : {
					"WALK": "walking"
				},
				"exit": "exited_standing" // event type for exit transition
			},
			"walking": {
				"entry": "entered_walking",
					"on" : {
						"STOP": "standing",
						"RUN": "running"
				}
			},
			"running": {
				"entry": "entered_running",
					"on" : {
						"WALK": "walking"
				}
			}
		}
	}`
	m := gomata.NewMachine(fsmJson)
	handler := func(state gomata.State) {
		fmt.Println("[Transition]: ", state.Value)
	}
	emitHandler := func(e gomata.Event) {
		fmt.Println("[Emit]: ", e.Type)
	}
	// OnTransition is a callback for state change
	m.OnTransition(&handler)
	err := m.Start() // [Transition]: standing; [Emit]: entered_standing
	if err != nil {
		fmt.Println("[ERR]: ", err)
	}
	m.Send("WALK") // [emit]: exited_standing > [Transition]: walking > [emit]: entered_walking
	m.Send("RUN")  // [Transition]: running > [emit]: entered_running

	// Start > standing > walking > running
}
```

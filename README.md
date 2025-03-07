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
	handler := func(e gomata.Event) {
		fmt.Println("[Transition]: ", e.Type)
	}
	m.OnTransition(&handler)
	err := m.Start() // [Transition]: entered_standing
	if err != nil {
		fmt.Println("[ERR]: ", err)
	}
	m.Send("WALK") // [Transition]: exited_standing > [Transition]: entered_walking
	m.Send("RUN")  // [Transition]: entered_running

	// Start > standing > walking > running > walking
}
```

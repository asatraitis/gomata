package main

import (
	"fmt"

	"github.com/asatraitis/gomata"
)

const machineConfig = `{
		"initial": "standing",
		"states": {
			"standing": {
				"entry": "entered_standing",
				"on" : {
					"WALK": "walking"
				},
				"exit": "exited_standing"
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

func main() {
	m := gomata.NewMachine(machineConfig)
	stateHandler := func(state gomata.State) {
		fmt.Println("[Transition]: ", state.Value)
	}
	m.OnTransition(&stateHandler)

	emitHandler := func(e gomata.Event) {
		fmt.Println("[Emit]: ", e.Type)
	}
	m.OnEmit(&emitHandler)
	fmt.Println("================= [START] =================")
	if err := m.Start(); err != nil {
		panic(err)
	}

	fmt.Println("================= [WALK] =================")
	m.Send("WALK")
	fmt.Println("================= [RUN] =================")
	m.Send("RUN")
}

package main

func main() {
	for _, e := range Educations {
		types := e.scheduleAvailableTypeUpdate()
		KBP.states = e.scheduleStatusUpdate(types[0])
		for _, state := range Educations[0].states {
			println(stateToBson(state))
			println(state.state)
		}
		e.scheduleUpdate(types[0])
	}
}

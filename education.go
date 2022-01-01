package main

type education struct {
	id                               int
	scheduleUpdateCronPattern        string
	primaryScheduleUpdateCronPattern string
	primaryCronStartTimePattern      string
	scheduleUpdate                   func(string, []StateInfo) []SubjectFull
	scheduleStatesUpdate             func(string) []StateInfo
	scheduleAvailableTypeUpdate      func() []string
	availableTypes                   []string
	states                           []StateInfo
}

package main

type education struct {
	educationPlaceId                 int
	scheduleUpdateCronPattern        string
	primaryScheduleUpdateCronPattern string
	primaryCronStartTimePattern      string
	scheduleUpdate                   func(string, []StateInfo) []Subject
	scheduleStatesUpdate             func(string) []StateInfo
	scheduleAvailableTypeUpdate      func() []string
	availableTypes                   []string
	states                           []StateInfo
}

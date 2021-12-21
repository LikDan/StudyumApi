package main

type education struct {
	educationPlaceId                        int
	scheduleUpdateCronePattern              string
	primaryScheduleUpdateCronePattern       string
	scheduleAvailableTypeUpdateCronePattern string
	scheduleUpdate                          func(string) []Subject
	scheduleStatusUpdate                    func(string) []StateInfo
	scheduleAvailableTypeUpdate             func() []string
	availableTypes                          []string
	states                                  []StateInfo
}

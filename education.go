package main

type education struct {
	educationPlaceId                        int
	scheduleUpdateCronePattern              string
	primaryScheduleUpdateCronePattern       string
	scheduleAvailableTypeUpdateCronePattern string
	scheduleUpdate                          func()
	scheduleStatusUpdate                    func()
	scheduleAvailableTypeUpdate             func() []string
	availableTypes                          interface{ string }
}

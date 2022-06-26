package models

type ScheduleStateInfo struct {
	State     State `bson:"status" json:"status"`
	WeekIndex int   `bson:"weekIndex" json:"weekIndex"`
	DayIndex  int   `bson:"dayIndex" json:"dayIndex"`
}

func GetScheduleStateInfoByIndexes(weekIndex, dayIndex int, states []*ScheduleStateInfo) *ScheduleStateInfo {
	for _, state := range states {
		if state.WeekIndex == weekIndex && state.DayIndex == dayIndex {
			return state
		}
	}

	return &ScheduleStateInfo{
		State:     NotUpdated,
		WeekIndex: weekIndex,
		DayIndex:  dayIndex,
	}
}

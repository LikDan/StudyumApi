package models

type State string

const (
	Updated    State = "UPDATED"
	NotUpdated State = "NOT_UPDATED"
)

type DayState struct {
	State        State `bson:"status" json:"status"`
	WeekIndex    int   `bson:"weekIndex" json:"weekIndex"`
	DayIndex     int   `bson:"dayIndex" json:"dayIndex"`
	StudyPlaceId int   `bson:"educationPlaceId" json:"-"`
}

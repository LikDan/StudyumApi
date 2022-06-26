package models

type IParserApp interface {
	Init(lesson Lesson)
	CommitUpdate()

	ScheduleUpdate(type_ *ScheduleTypeInfo) []*Lesson
	GeneralScheduleUpdate(type_ *ScheduleTypeInfo) []*GeneralLesson
	ScheduleTypesUpdate() []*ScheduleTypeInfo
	JournalUpdate(user *ParseJournalUser) []*Mark

	GetName() string
	GetStudyPlaceId() int
	GetUpdateCronPattern() string
}

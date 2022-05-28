package models

type StudyPlace struct {
	Id         int    `json:"id" bson:"_id"`
	WeeksCount int    `json:"weeksCount" bson:"weeksCount"`
	DaysCount  int    `json:"daysCount" bson:"daysCount"`
	Name       string `json:"name" bson:"name"`
}

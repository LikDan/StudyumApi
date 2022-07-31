package models

type JournalRow struct {
	Id       string    `json:"id" bson:"_id"`
	Title    string    `json:"title" bson:"title"`
	UserType string    `json:"userType" bson:"userType"`
	Lessons  []*Lesson `json:"lessons" bson:"lessons"`
}

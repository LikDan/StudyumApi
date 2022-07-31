package models

type JournalInfo struct {
	Editable     bool   `json:"editable" bson:"editable"`
	StudyPlaceId int    `json:"studyPlaceId" bson:"studyPlaceId"`
	Group        string `json:"group" bson:"group"`
	Teacher      string `json:"teacher" bson:"teacher"`
	Subject      string `json:"subject" bson:"subject"`
}

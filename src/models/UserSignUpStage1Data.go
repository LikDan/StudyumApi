package models

type UserSignUpStage1Data struct {
	StudyPlaceId int    `bson:"studyPlaceId" json:"studyPlaceId"`
	Type         string `bson:"type" json:"type"`
	TypeName     string `bson:"typeName" json:"typeName"`
}

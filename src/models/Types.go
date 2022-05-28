package models

type Types struct {
	Groups   []string `json:"groups" bson:"groups"`
	Teachers []string `json:"teachers" bson:"teachers"`
	Subjects []string `json:"subjects" bson:"subjects"`
	Rooms    []string `json:"rooms" bson:"rooms"`
}

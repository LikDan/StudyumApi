package main

import "go.mongodb.org/mongo-driver/bson"

type SubjectFull struct {
	subject          string
	teacher          string
	group            string
	room             string
	columnIndex      int
	rowIndex         int
	weekIndex        int
	type_            string
	educationPlaceId int
}

func subjectToBson(subject SubjectFull) bson.D {
	return bson.D{
		{"columnIndex", subject.columnIndex},
		{"rowIndex", subject.rowIndex},
		{"weekIndex", subject.weekIndex},
		{"subject", subject.subject},
		{"teacher", subject.teacher},
		{"group", subject.group},
		{"room", subject.room},
		{"type", subject.type_},
		{"educationPlaceId", subject.educationPlaceId},
	}
}

func subjectToBsonWithoutType(subject SubjectFull) bson.D {
	return bson.D{
		{"columnIndex", subject.columnIndex},
		{"rowIndex", subject.rowIndex},
		{"weekIndex", subject.weekIndex},
		{"subject", subject.subject},
		{"teacher", subject.teacher},
		{"group", subject.group},
		{"room", subject.room},
		{"educationPlaceId", subject.educationPlaceId},
	}
}

package main

import (
	"strconv"
	"strings"
)

type Lesson struct {
	subjects    []Subject
	columnIndex int32
	rowIndex    int32
	weekIndex   int32
}

type Subject struct {
	subject string
	teacher string
	group   string
	room    string
	type_   string
}

func (l Lesson) toJson() string {
	isStay := "true"
	var subjects []string
	for _, subject := range l.subjects {
		if subject.type_ != "STAY" && isStay == "true" {
			isStay = "false"
		}
		subjects = append(subjects, subject.toJson())
	}

	return "{\"weekIndex\": " + strconv.Itoa(int(l.weekIndex)) +
		", \"columnIndex\": " + strconv.Itoa(int(l.columnIndex)) +
		", \"rowIndex\": " + strconv.Itoa(int(l.rowIndex)) +
		", \"isStay\": " + isStay +
		", \"subjects\": [" + strings.Join(subjects, ", ") + "]}"
}

func (s Subject) toJson() string {
	return "{\"subject\": \"" + s.subject +
		"\", \"teacher\": \"" + s.teacher +
		"\", \"group\": \"" + s.group +
		"\", \"room\": \"" + s.room +
		"\", \"type\": \"" + s.type_ + "\"}"
}

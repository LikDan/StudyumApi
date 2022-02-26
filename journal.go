package main

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func getTeacherJournalSubjects(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")

	var subjects []SubjectFull

	find, err := subjectsCollection.Find(nil, bson.M{"teacher": user.FullName, "group": group, "subject": subject})
	err = find.All(nil, &subjects)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	if subjects == nil {
		errorMessage(ctx, "no such subjects")
		return
	}

	ctx.JSON(200, subjects)
}

func getTeacherJournalTypes(ctx *gin.Context) {
	user, err := getUserFromDbViaCookies(ctx)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	find, err := generalSubjectsCollection.Find(nil, bson.M{"teacher": user.FullName})
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	var subjects []SubjectFull
	err = find.All(nil, &subjects)
	if checkError(err) {
		errorMessage(ctx, err.Error())
		return
	}

	var types []JournalTeacherType

	for _, subject := range subjects {
		type_ := JournalTeacherType{
			Teacher: subject.Teacher,
			Subject: subject.Subject,
			Group:   subject.Group,
		}

		if sliceContains(types, type_) {
			continue
		}

		types = append(types, type_)
	}

	ctx.JSON(200, types)
}

func addMark(ctx *gin.Context) {

}

type JournalTeacherType struct {
	Teacher string `json:"teacher"`
	Subject string `json:"subject"`
	Group   string `json:"group"`
}

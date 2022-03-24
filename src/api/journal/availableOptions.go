package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	h "studyium/api"
	"studyium/api/schedule"
	userApi "studyium/api/user"
	"studyium/db"
)

func getAvailableOptions(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	var filter bson.M
	if user.Type == "teacher" {
		filter = bson.M{"teacher": user.FullName}
	} else {
		filter = bson.M{"group": user.Name}
	}

	find, err := db.GeneralSubjectsCollection.Find(nil, filter)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	var subjects []schedule.SubjectFull
	err = find.All(nil, &subjects)
	if h.CheckError(err, h.WARNING) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	var types []JournalTeacherType

	for _, subject := range subjects {
		type_ := JournalTeacherType{
			Teacher:  subject.Teacher,
			Subject:  subject.Subject,
			Group:    subject.Group,
			Editable: h.SliceContains(user.Permissions, "editJournal"),
		}

		if h.SliceContains(types, type_) {
			continue
		}

		types = append(types, type_)
	}

	ctx.JSON(200, types)
}

package journal

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	h "studyium/src/api"
	"studyium/src/api/schedule"
	userApi "studyium/src/api/user"
	"studyium/src/db"
)

func getAvailableOptions(ctx *gin.Context) {
	user, err := userApi.GetUserFromDbViaCookies(ctx)
	if h.CheckError(err, h.UNDEFINED) {
		h.ErrorMessage(ctx, err.Error())
		return
	}

	if user.Type == "group" {
		ctx.JSON(200, []AvailableOption{{
			Teacher:  "",
			Subject:  "",
			Group:    user.Name,
			Editable: false,
		}})
		return
	}

	find, err := db.GeneralSubjectsCollection.Find(nil, bson.M{"teacher": user.FullName})
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

	var types []AvailableOption

	for _, subject := range subjects {
		type_ := AvailableOption{
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

type AvailableOption struct {
	Teacher  string `json:"teacher"`
	Subject  string `json:"subject"`
	Group    string `json:"group"`
	Editable bool   `json:"editable"`
}

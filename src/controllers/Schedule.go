package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/api"
	"studyum/src/api/parser"
	"studyum/src/api/parser/studyPlace"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/utils"
)

func GetSchedule(ctx *gin.Context) {
	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	if utils.CheckEmptyAndResponse(ctx, models.BindErrorStr("provide valid params", 400, api.UNDEFINED), type_, typeName) {
		return
	}

	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	var schedule models.Schedule
	if err := db.GetSchedule(user.StudyPlaceId, type_, typeName, &schedule); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, schedule)
}

func GetMySchedule(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	var schedule models.Schedule
	if err := db.GetSchedule(user.StudyPlaceId, user.Type, user.TypeName, &schedule); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, schedule)
}

func GetScheduleTypes(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	types := models.Types{
		Groups:   db.GetScheduleType(user.StudyPlaceId, "group"),
		Teachers: db.GetScheduleType(user.StudyPlaceId, "teacher"),
		Subjects: db.GetScheduleType(user.StudyPlaceId, "subject"),
		Rooms:    db.GetScheduleType(user.StudyPlaceId, "room"),
	}

	ctx.JSON(200, types)
}

func Update(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user, "editSchedule"); err.CheckAndResponse(ctx) {
		return
	}

	var education *studyPlace.Education = nil
	for _, education_ := range parser.Educations {
		if education_.Id == user.StudyPlaceId {
			education = education_
			break
		}
	}
	if education == nil {
		models.BindErrorStr("not authorized", 401, api.UNDEFINED).CheckAndResponse(ctx)
	}

	parser.UpdateDbSchedule(education)
}

func AddLesson(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user, "editSchedule"); err.CheckAndResponse(ctx) {
		return
	}

	var subject models.Lesson
	if err := ctx.BindJSON(&subject); models.BindError(err, 400, api.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := db.AddLesson(&subject, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, subject)
}

func DeleteLesson(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user, "editSchedule"); err.CheckAndResponse(ctx) {
		return
	}

	idHex := ctx.Param("id")
	if !primitive.IsValidObjectID(idHex) {
		models.BindErrorStr("provide valid id", 400, api.UNDEFINED).CheckAndResponse(ctx)
	}

	id, _ := primitive.ObjectIDFromHex(idHex)
	if err := db.DeleteLesson(id, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, id)
}

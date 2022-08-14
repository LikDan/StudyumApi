package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/db"
	"studyum/src/models"
	"studyum/src/utils"
)

func GetJournalAvailableOptions(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if user.Type == "group" {
		ctx.JSON(200, []models.JournalAvailableOption{{
			Teacher:  "",
			Subject:  "",
			Group:    user.TypeName,
			Editable: false,
		}})
		return
	}

	options, err := db.GetAvailableOptions(ctx, user.Name, utils.SliceContains(user.Permissions, "editJournal"))
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, options)
}

func GetJournal(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if !utils.CheckNotEmpty(ctx.Param("group"), ctx.Param("subject"), ctx.Param("teacher")) {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	var journal models.Journal
	if err := db.GetJournal(ctx, &journal, ctx.Param("group"), ctx.Param("subject"), user.TypeName, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, journal)
}

func GetUserJournal(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	var journal models.Journal
	if err := db.GetStudentJournal(ctx, &journal, user.Id, user.TypeName, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, journal)
}

func AddMark(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if !utils.SliceContains(user.Permissions, "editJournal") {
		models.BindErrorStr("no permission", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	var mark models.Mark
	if err := ctx.BindJSON(&mark); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if mark.Mark == "" || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	if err := db.AddMark(&mark); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := db.GetLessonById(ctx, mark.UserId, mark.LessonId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

func GetMark(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")
	teacher := user.Name

	if group == "" || subject == "" || userIdHex == "" {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	userId, err_ := primitive.ObjectIDFromHex(userIdHex)
	if models.BindError(err_, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	lessons, err := db.GetLessons(ctx, userId, group, teacher, subject, user.StudyPlaceId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lessons)
}

func UpdateMark(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if !utils.SliceContains(user.Permissions, "editJournal") {
		models.BindErrorStr("no permission", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	var mark models.Mark
	if err := ctx.BindJSON(&mark); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if mark.Mark == "" || mark.Id.IsZero() || mark.UserId.IsZero() || mark.LessonId.IsZero() {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED)
		return
	}

	if err := db.UpdateMark(&mark); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := db.GetLessonById(ctx, mark.UserId, mark.LessonId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

func DeleteMark(ctx *gin.Context) {
	var user models.User
	if err := AuthUserViaContext(ctx, &user); err.CheckAndResponse(ctx) {
		return
	}

	if !utils.SliceContains(user.Permissions, "editJournal") {
		models.BindErrorStr("no permission", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	markId := ctx.Query("markId")
	userIdHex := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")

	if markId == "" || userIdHex == "" || subjectId == "" {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	userId, err_ := primitive.ObjectIDFromHex(userIdHex)
	if models.BindError(err_, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	markObjectId, err_ := primitive.ObjectIDFromHex(markId)
	if models.BindError(err_, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	subjectObjectId, err_ := primitive.ObjectIDFromHex(subjectId)
	if models.BindError(err_, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := db.DeleteMark(markObjectId, subjectObjectId); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := db.GetLessonById(ctx, userId, subjectObjectId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

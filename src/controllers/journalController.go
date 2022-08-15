package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/repositories"
	"studyum/src/utils"
)

type JournalController struct {
	repository repositories.IJournalRepository
}

func NewJournalController(repository repositories.IJournalRepository) *JournalController {
	return &JournalController{repository: repository}
}

func (j *JournalController) GetJournalAvailableOptions(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	if user.Type == "group" {
		ctx.JSON(200, []models.JournalAvailableOption{{
			Teacher:  "",
			Subject:  "",
			Group:    user.TypeName,
			Editable: false,
		}})
		return
	}

	options, err := j.repository.GetAvailableOptions(ctx, user.Name, utils.SliceContains(user.Permissions, "editJournal"))
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, options)
}

func (j *JournalController) GetJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	if !utils.CheckNotEmpty(ctx.Param("group"), ctx.Param("subject"), ctx.Param("teacher")) {
		models.BindErrorStr("provide valid params", 400, models.UNDEFINED).CheckAndResponse(ctx)
		return
	}

	var journal models.Journal
	if err := j.repository.GetJournal(ctx, &journal, ctx.Param("group"), ctx.Param("subject"), user.TypeName, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, journal)
}

func (j *JournalController) GetUserJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var journal models.Journal
	if err := j.repository.GetStudentJournal(ctx, &journal, user.Id, user.TypeName, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, journal)
}

func (j *JournalController) AddMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	if err := j.repository.AddMark(ctx, &mark); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

func (j *JournalController) GetMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	lessons, err := j.repository.GetLessons(ctx, userId, group, teacher, subject, user.StudyPlaceId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lessons)
}

func (j *JournalController) UpdateMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	if err := j.repository.UpdateMark(ctx, &mark); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := j.repository.GetLessonById(ctx, mark.UserId, mark.LessonId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

func (j *JournalController) DeleteMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

	if err := j.repository.DeleteMark(ctx, markObjectId, subjectObjectId); err.CheckAndResponse(ctx) {
		return
	}

	lesson, err := j.repository.GetLessonById(ctx, userId, subjectObjectId)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, lesson)
}

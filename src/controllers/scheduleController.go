package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"studyum/src/models"
	"studyum/src/parser"
	"studyum/src/repositories"
	"studyum/src/utils"
)

type ScheduleController struct {
	repository repositories.IScheduleRepository
}

func NewScheduleController(repository repositories.IScheduleRepository) *ScheduleController {
	return &ScheduleController{repository: repository}
}

func (s *ScheduleController) GetSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	if utils.CheckEmptyAndResponse(ctx, models.BindErrorStr("provide valid params", 400, models.UNDEFINED), type_, typeName) {
		return
	}

	var schedule models.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, type_, typeName, &schedule); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, schedule)
}

func (s *ScheduleController) GetMySchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var schedule models.Schedule
	if err := s.repository.GetSchedule(ctx, user.StudyPlaceId, user.Type, user.TypeName, &schedule); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, schedule)
}

func (s *ScheduleController) GetScheduleTypes(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	types := models.Types{
		Groups:   s.repository.GetScheduleType(ctx, user.StudyPlaceId, "group"),
		Teachers: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "teacher"),
		Subjects: s.repository.GetScheduleType(ctx, user.StudyPlaceId, "subject"),
		Rooms:    s.repository.GetScheduleType(ctx, user.StudyPlaceId, "room"),
	}

	ctx.JSON(200, types)
}

func (s *ScheduleController) UpdateSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var app models.IParserApp
	if err := parser.GetAppByStudyPlaceId(user.StudyPlaceId, &app); err.CheckAndResponse(ctx) {
		return
	}

	parser.Update(app)
	ctx.JSON(200, "updated")
}

func (s *ScheduleController) UpdateGeneralSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var app models.IParserApp
	if err := parser.GetAppByStudyPlaceId(user.StudyPlaceId, &app); err.CheckAndResponse(ctx) {
		return
	}

	parser.UpdateGeneralSchedule(app)
	ctx.JSON(200, "updated")
}

func (s *ScheduleController) AddLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var subject models.Lesson
	if err := ctx.BindJSON(&subject); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := s.repository.AddLesson(ctx, &subject, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, subject)
}

func (s *ScheduleController) UpdateLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var subject models.Lesson
	if err := ctx.BindJSON(&subject); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	if err := s.repository.UpdateLesson(ctx, &subject, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, subject)
}

func (s *ScheduleController) DeleteLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	idHex := ctx.Param("id")
	if !primitive.IsValidObjectID(idHex) {
		models.BindErrorStr("provide valid id", 400, models.UNDEFINED).CheckAndResponse(ctx)
	}

	id, _ := primitive.ObjectIDFromHex(idHex)
	if err := s.repository.DeleteLesson(ctx, id, user.StudyPlaceId); err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(200, id)
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/dto"
	"studyum/internal/utils"
	"time"
)

type ScheduleHandler interface {
	GetScheduleTypes(ctx *gin.Context)

	GetSchedule(ctx *gin.Context)
	GetUserSchedule(ctx *gin.Context)

	AddLesson(ctx *gin.Context)
	UpdateLesson(ctx *gin.Context)
	DeleteLesson(ctx *gin.Context)

	AddGeneralLessons(ctx *gin.Context)

	AddLessons(ctx *gin.Context)
	RemoveLessonsBetweenDates(ctx *gin.Context)

	SaveCurrentScheduleAsGeneral(ctx *gin.Context)
	SaveGeneralScheduleAsCurrent(ctx *gin.Context)
}

type scheduleHandler struct {
	Handler

	controller controllers.ScheduleController

	Group *gin.RouterGroup
}

func NewScheduleHandler(authHandler Handler, controller controllers.ScheduleController, group *gin.RouterGroup) ScheduleHandler {
	h := &scheduleHandler{Handler: authHandler, controller: controller, Group: group}

	group.GET("getTypes", h.User(), h.GetScheduleTypes)
	group.GET(":type/:name", h.User(), h.GetSchedule)
	group.GET("", h.Auth(), h.GetUserSchedule)

	group.POST("", h.Auth("editSchedule"), h.AddLesson)
	group.PUT("", h.Auth("editJournal"), h.UpdateLesson)
	group.DELETE(":id", h.Auth("editSchedule"), h.DeleteLesson)
	group.DELETE("between/:startDate/:endDate", h.Auth("editSchedule"), h.RemoveLessonsBetweenDates)

	group.POST("/list", h.Auth("editSchedule"), h.AddLessons)
	group.POST("/general/list", h.Auth("editSchedule"), h.AddGeneralLessons)

	group.POST("/makeGeneral", h.Auth("editSchedule"), h.SaveCurrentScheduleAsGeneral)
	group.POST("/makeCurrent/:date", h.Auth("editSchedule"), h.SaveGeneralScheduleAsCurrent)

	return h
}

func (s *scheduleHandler) GetSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")

	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	schedule, err := s.controller.GetSchedule(ctx, studyPlaceID, type_, typeName, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *scheduleHandler) GetUserSchedule(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	schedule, err := s.controller.GetUserSchedule(ctx, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *scheduleHandler) GetScheduleTypes(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	id := ctx.Query("id")
	types := s.controller.GetScheduleTypes(ctx, user, id)

	ctx.JSON(http.StatusOK, types)
}

func (s *scheduleHandler) AddLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lessonDTO dto.AddLessonDTO
	if err := ctx.BindJSON(&lessonDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := s.controller.AddLesson(ctx, lessonDTO, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *scheduleHandler) AddLessons(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lessonsDTO []dto.AddLessonDTO
	if err := ctx.BindJSON(&lessonsDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lessons, err := s.controller.AddLessons(ctx, user, lessonsDTO)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (s *scheduleHandler) AddGeneralLessons(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lessonsDTO []dto.AddGeneralLessonDTO
	if err := ctx.BindJSON(&lessonsDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lessons, err := s.controller.AddGeneralLessons(ctx, user, lessonsDTO)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (s *scheduleHandler) UpdateLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var lesson dto.UpdateLessonDTO
	if err := ctx.BindJSON(&lesson); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err := s.controller.UpdateLesson(ctx, lesson, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *scheduleHandler) DeleteLesson(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)
	id := ctx.Param("id")

	err := s.controller.DeleteLesson(ctx, id, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}

func (s *scheduleHandler) SaveCurrentScheduleAsGeneral(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	type_ := ctx.Query("type")
	typeName := ctx.Query("typeName")

	err := s.controller.SaveCurrentScheduleAsGeneral(ctx, user, type_, typeName)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

func (s *scheduleHandler) SaveGeneralScheduleAsCurrent(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	date_ := ctx.Param("date")
	date, err := time.Parse(time.RFC3339, date_)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = s.controller.SaveGeneralScheduleAsCurrent(ctx, user, date); err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

func (s *scheduleHandler) RemoveLessonsBetweenDates(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	startDate, err := time.Parse(time.RFC3339, ctx.Param("startDate"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	endDate, err := time.Parse(time.RFC3339, ctx.Param("endDate"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = s.controller.RemoveLessonBetweenDates(ctx, user, startDate, endDate); err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, "removed")
}

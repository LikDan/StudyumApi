package schedule

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/global"
	"time"
)

type Handler interface {
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

type handler struct {
	global.Handler

	controller Controller

	Group *gin.RouterGroup
}

func NewScheduleHandler(authHandler global.Handler, controller Controller, group *gin.RouterGroup) Handler {
	h := &handler{Handler: authHandler, controller: controller, Group: group}

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

func (s *handler) GetSchedule(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

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

func (s *handler) GetUserSchedule(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	schedule, err := s.controller.GetUserSchedule(ctx, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *handler) GetScheduleTypes(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	id := ctx.Query("id")
	types := s.controller.GetScheduleTypes(ctx, user, id)

	ctx.JSON(http.StatusOK, types)
}

func (s *handler) AddLesson(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	var lessonDTO AddLessonDTO
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

func (s *handler) AddLessons(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	var lessonsDTO []AddLessonDTO
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

func (s *handler) AddGeneralLessons(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	var lessonsDTO []AddGeneralLessonDTO
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

func (s *handler) UpdateLesson(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	var lesson UpdateLessonDTO
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

func (s *handler) DeleteLesson(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)
	id := ctx.Param("id")

	err := s.controller.DeleteLesson(ctx, id, user)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}

func (s *handler) SaveCurrentScheduleAsGeneral(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

	type_ := ctx.Query("type")
	typeName := ctx.Query("typeName")

	err := s.controller.SaveCurrentScheduleAsGeneral(ctx, user, type_, typeName)
	if err != nil {
		s.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

func (s *handler) SaveGeneralScheduleAsCurrent(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

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

func (s *handler) RemoveLessonsBetweenDates(ctx *gin.Context) {
	user := s.GetUserViaCtx(ctx)

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

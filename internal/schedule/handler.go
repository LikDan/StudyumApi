package schedule

import (
	"github.com/gin-gonic/gin"
	"net/http"
	auth "studyum/internal/auth/handlers"
	"time"
)

type Handler interface {
	GetScheduleTypes(ctx *gin.Context)

	GetSchedule(ctx *gin.Context)
	GetUserSchedule(ctx *gin.Context)

	GetGeneralSchedule(ctx *gin.Context)
	GetGeneralUserSchedule(ctx *gin.Context)

	AddLesson(ctx *gin.Context)
	UpdateLesson(ctx *gin.Context)
	DeleteLesson(ctx *gin.Context)
	GetLessonByID(ctx *gin.Context)

	AddGeneralLessons(ctx *gin.Context)

	AddLessons(ctx *gin.Context)
	RemoveLessonsBetweenDates(ctx *gin.Context)

	SaveCurrentScheduleAsGeneral(ctx *gin.Context)
	SaveGeneralScheduleAsCurrent(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller Controller

	Group *gin.RouterGroup
}

func NewScheduleHandler(middleware auth.Middleware, controller Controller, group *gin.RouterGroup) Handler {
	h := &handler{Middleware: middleware, controller: controller, Group: group}

	group.GET("getTypes", h.TryAuth(), h.GetScheduleTypes)
	group.GET(":type/:name", h.TryAuth(), h.GetSchedule)
	group.GET("", h.MemberAuth(), h.GetUserSchedule)

	group.GET("general/:type/:name", h.MemberAuth(), h.GetGeneralSchedule)
	group.GET("general", h.MemberAuth(), h.GetGeneralUserSchedule)

	group.POST("", h.MemberAuth("editSchedule"), h.AddLesson)
	group.PUT("", h.MemberAuth("editJournal"), h.UpdateLesson)
	group.DELETE(":id", h.MemberAuth("editSchedule"), h.DeleteLesson)
	group.GET("lessons/:id", h.MemberAuth(), h.GetLessonByID)
	group.DELETE("between/:startDate/:endDate", h.MemberAuth("editSchedule"), h.RemoveLessonsBetweenDates)

	group.POST("/list", h.MemberAuth("editSchedule"), h.AddLessons)
	group.POST("/general/list", h.MemberAuth("editSchedule"), h.AddGeneralLessons)

	group.POST("/makeGeneral", h.MemberAuth("editSchedule"), h.SaveCurrentScheduleAsGeneral)
	group.POST("/makeCurrent/:date", h.MemberAuth("editSchedule"), h.SaveGeneralScheduleAsCurrent)

	return h
}

func (s *handler) GetSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")

	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	schedule, err := s.controller.GetSchedule(ctx, studyPlaceID, type_, typeName, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *handler) GetUserSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	schedule, err := s.controller.GetUserSchedule(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *handler) GetGeneralSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")

	type_ := ctx.Param("type")
	typeName := ctx.Param("name")

	schedule, err := s.controller.GetGeneralSchedule(ctx, studyPlaceID, type_, typeName, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *handler) GetGeneralUserSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	schedule, err := s.controller.GetGeneralUserSchedule(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

func (s *handler) GetScheduleTypes(ctx *gin.Context) {
	user := s.GetUser(ctx)

	id := ctx.Query("id")
	types := s.controller.GetScheduleTypes(ctx, user, id)

	ctx.JSON(http.StatusOK, types)
}

func (s *handler) GetLessonByID(ctx *gin.Context) {
	user := s.GetUser(ctx)

	id := ctx.Param("id")
	if ctx.Query("type") == "date" {
		lesson, err := s.controller.GetLessonsByDateAndID(ctx, user, id)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, lesson)
		return
	}

	lesson, err := s.controller.GetLessonByID(ctx, user, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *handler) AddLesson(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonDTO AddLessonDTO
	if err := ctx.BindJSON(&lessonDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := s.controller.AddLesson(ctx, lessonDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *handler) AddLessons(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonsDTO []AddLessonDTO
	if err := ctx.BindJSON(&lessonsDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lessons, err := s.controller.AddLessons(ctx, user, lessonsDTO)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (s *handler) AddGeneralLessons(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonsDTO []AddGeneralLessonDTO
	if err := ctx.BindJSON(&lessonsDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lessons, err := s.controller.AddGeneralLessons(ctx, user, lessonsDTO)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (s *handler) UpdateLesson(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lesson UpdateLessonDTO
	if err := ctx.BindJSON(&lesson); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err := s.controller.UpdateLesson(ctx, lesson, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (s *handler) DeleteLesson(ctx *gin.Context) {
	user := s.GetUser(ctx)
	id := ctx.Param("id")

	err := s.controller.DeleteLesson(ctx, id, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, id)
}

func (s *handler) SaveCurrentScheduleAsGeneral(ctx *gin.Context) {
	user := s.GetUser(ctx)

	type_ := ctx.Query("type")
	typeName := ctx.Query("typeName")

	err := s.controller.SaveCurrentScheduleAsGeneral(ctx, user, type_, typeName)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

func (s *handler) SaveGeneralScheduleAsCurrent(ctx *gin.Context) {
	user := s.GetUser(ctx)

	date_ := ctx.Param("date")
	date, err := time.Parse(time.RFC3339, date_)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err = s.controller.SaveGeneralScheduleAsCurrent(ctx, user, date); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

func (s *handler) RemoveLessonsBetweenDates(ctx *gin.Context) {
	user := s.GetUser(ctx)

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
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "removed")
}

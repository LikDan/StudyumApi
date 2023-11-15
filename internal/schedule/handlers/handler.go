package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/schedule/controllers"
	"studyum/internal/schedule/dto"
	"time"
)

type Handler interface {
	GetSchedule(ctx *gin.Context)
	GetUserSchedule(ctx *gin.Context)

	GetGeneralSchedule(ctx *gin.Context)
	GetGeneralUserSchedule(ctx *gin.Context)

	GetGeneralLessonsList(ctx *gin.Context)

	GetScheduleTypes(ctx *gin.Context)

	GetLessonByID(ctx *gin.Context)
	AddLessons(ctx *gin.Context)
	AddScheduleInfo(ctx *gin.Context)
	AddLesson(ctx *gin.Context)
	UpdateLesson(ctx *gin.Context)
	DeleteLesson(ctx *gin.Context)
	RemoveLessonsBetweenDates(ctx *gin.Context)

	AddGeneralLessons(ctx *gin.Context)

	SaveCurrentScheduleAsGeneral(ctx *gin.Context)
	SaveGeneralScheduleAsCurrent(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller controllers.Controller

	Group *gin.RouterGroup
}

func NewScheduleHandler(middleware auth.Middleware, controller controllers.Controller, group *gin.RouterGroup) Handler {
	h := &handler{Middleware: middleware, controller: controller, Group: group}

	//group.GET(":type/:name", h.TryAuth(), h.GetSchedule)
	group.GET("", h.TryAuth(), h.GetUserSchedule)

	group.GET("general/:type/:name", h.MemberAuth(), h.GetGeneralSchedule)
	group.GET("general", h.MemberAuth(), h.GetGeneralUserSchedule)

	group.GET("types", h.TryAuth(), h.GetScheduleTypes) //todo change endpoint to types

	group.POST("/info", h.MemberAuth("editSchedule"), h.AddScheduleInfo)

	group.GET("lessons/:id", h.MemberAuth(), h.GetLessonByID) //todo change endpoint to :id
	group.POST("/list", h.MemberAuth("editSchedule"), h.AddLessons)
	group.POST("lessons", h.MemberAuth("editSchedule"), h.AddLesson)
	group.PUT("", h.MemberAuth("editJournal"), h.UpdateLesson)
	group.DELETE(":id", h.MemberAuth("editSchedule"), h.DeleteLesson)
	group.DELETE("between/:startDate/:endDate", h.MemberAuth("editSchedule"), h.RemoveLessonsBetweenDates)

	group.POST("/general/list", h.MemberAuth("editSchedule"), h.AddGeneralLessons)
	group.GET("/general/list", h.MemberAuth(), h.GetGeneralLessonsList)

	group.POST("/makeGeneral", h.MemberAuth("editSchedule"), h.SaveCurrentScheduleAsGeneral)
	group.POST("/makeCurrent/:date", h.MemberAuth("editSchedule"), h.SaveGeneralScheduleAsCurrent)

	return h
}

// GetSchedule godoc
// @Param type path string true "Role"
// @Param name path string true "RoleName"
// @Router /{type}/{name} [get]
func (s *handler) GetSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")
	startDateStr := ctx.Query("startDate")
	startDate, _ := time.Parse(time.RFC3339, startDateStr)
	endDateStr := ctx.Query("endDate")
	endDate, _ := time.Parse(time.RFC3339, endDateStr)

	role := ctx.Param("type")
	roleName := ctx.Param("name")

	schedule, err := s.controller.GetSchedule(ctx, user, studyPlaceID, role, roleName, startDate, endDate, false)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// GetUserSchedule godoc
// @Router / [get]
func (s *handler) GetUserSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	startDateStr := ctx.Query("startDate")
	startDate, err := time.Parse(time.RFC3339, startDateStr)
	endDateStr := ctx.Query("endDate")
	endDate, _ := time.Parse(time.RFC3339, endDateStr)

	studyPlaceID := ctx.Query("studyPlaceID")
	if studyPlaceID == "" {
		studyPlaceID = user.StudyPlaceInfo.ID.Hex()
	}

	type_ := ctx.Query("type")
	typename := ctx.Query("typename")
	isGeneral := ctx.Query("general") == "true"

	if type_ != "" && typename != "" {
		schedule, err := s.controller.GetSchedule(ctx, user, studyPlaceID, type_, typename, startDate, endDate, isGeneral)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, schedule)
		return
	}

	schedule, err := s.controller.GetUserSchedule(ctx, user, startDate, endDate, isGeneral)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// GetGeneralSchedule godoc
// @Param type path string true "Type"
// @Param name path string true "RoleName"
// @Router /general/{type}/{name} [get]
func (s *handler) GetGeneralSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")

	role := ctx.Param("type")
	roleName := ctx.Param("name")

	startDateStr := ctx.Query("startDate")
	startDate, _ := time.Parse(time.RFC3339, startDateStr)
	endDateStr := ctx.Query("endDate")
	endDate, _ := time.Parse(time.RFC3339, endDateStr)

	schedule, err := s.controller.GetGeneralSchedule(ctx, user, studyPlaceID, role, roleName, startDate, endDate)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// GetGeneralLessonsList godoc
// @Router /general/list [get]
func (s *handler) GetGeneralLessonsList(ctx *gin.Context) {
	user := s.GetUser(ctx)

	studyPlaceID := ctx.Query("studyPlaceID")
	weekIndexStr := ctx.Query("weekIndex")
	dayIndexStr := ctx.Query("dayIndex")

	var weekIndex *int = nil
	var dayIndex *int = nil

	if weekIndexStr != "" {
		i, _ := strconv.Atoi(weekIndexStr)
		weekIndex = &i
	}

	if dayIndexStr != "" {
		i, _ := strconv.Atoi(dayIndexStr)
		dayIndex = &i
	}

	lessons, err := s.controller.GetGeneralLessons(ctx, user, studyPlaceID, weekIndex, dayIndex)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

// GetGeneralUserSchedule godoc
// @Router /general [get]
func (s *handler) GetGeneralUserSchedule(ctx *gin.Context) {
	user := s.GetUser(ctx)

	startDateStr := ctx.Query("startDate")
	startDate, _ := time.Parse(time.RFC3339, startDateStr)
	endDateStr := ctx.Query("endDate")
	endDate, _ := time.Parse(time.RFC3339, endDateStr)

	schedule, err := s.controller.GetGeneralUserSchedule(ctx, user, startDate, endDate)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// GetScheduleTypes godoc
// @Router /getTypes [get]
func (s *handler) GetScheduleTypes(ctx *gin.Context) {
	user := s.GetUser(ctx)

	id := ctx.Query("studyPlaceID")
	types := s.controller.GetScheduleTypes(ctx, user, id)

	ctx.JSON(http.StatusOK, types)
}

// GetLessonByID godoc
// @Param id path string true "Lesson ID"
// @Router /lessons/{id} [get]
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

// AddLessons godoc
// @Router /list [post]
func (s *handler) AddLessons(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonsDTO []dto.AddLessonDTO
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

// AddLesson godoc
// @Router / [post]
func (s *handler) AddLesson(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonDTO dto.AddLessonDTO
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

// AddScheduleInfo godoc
// @Router /info [post]
func (s *handler) AddScheduleInfo(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var infoDTO dto.AddScheduleInfoDTO
	if err := ctx.BindJSON(&infoDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	info, err := s.controller.AddScheduleInfo(ctx, infoDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, info)
}

// UpdateLesson godoc
// @Router / [put]
func (s *handler) UpdateLesson(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lesson dto.UpdateLessonDTO
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

// DeleteLesson godoc
// @Param id path string true "Lesson ID"
// @Router / [delete]
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

// RemoveLessonsBetweenDates godoc
// @Param startDate path string true "From date"
// @Param endDate path string true "To date"
// @Router /between/{startDate}/{endDate} [delete]
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

// AddGeneralLessons godoc
// @Router /general/list [post]
func (s *handler) AddGeneralLessons(ctx *gin.Context) {
	user := s.GetUser(ctx)

	var lessonsDTO []dto.AddGeneralLessonDTO
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

// SaveCurrentScheduleAsGeneral godoc
// @Router /makeGeneral [post]
func (s *handler) SaveCurrentScheduleAsGeneral(ctx *gin.Context) {
	user := s.GetUser(ctx)

	role := ctx.Query("type")
	roleName := ctx.Query("roleName")

	err := s.controller.SaveCurrentScheduleAsGeneral(ctx, user, role, roleName)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, "successful")
}

// SaveGeneralScheduleAsCurrent godoc
// @Param date path string true "Date"
// @Router /makeCurrent/:date [post]
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

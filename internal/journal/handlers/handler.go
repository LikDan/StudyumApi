package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	auth "studyum/internal/auth/handlers"
	"studyum/internal/journal/controllers"
	"studyum/internal/journal/dtos"
	"studyum/internal/journal/entities"
)

type Handler interface {
	GenerateMarks(ctx *gin.Context)
	GenerateAbsences(ctx *gin.Context)

	GetJournalAvailableOptions(ctx *gin.Context)

	GetJournal(ctx *gin.Context)
	GetUserJournal(ctx *gin.Context)

	AddMarks(ctx *gin.Context)
	AddMark(ctx *gin.Context)
	UpdateMark(ctx *gin.Context)
	DeleteMark(ctx *gin.Context)

	AddAbsences(ctx *gin.Context)
	AddAbsence(ctx *gin.Context)
	UpdateAbsence(ctx *gin.Context)
	DeleteAbsence(ctx *gin.Context)

	GetLesson(ctx *gin.Context)
}

type handler struct {
	auth.Middleware

	controller        controllers.Controller
	journalController controllers.Journal

	Group *gin.RouterGroup
}

func NewJournalHandler(middleware auth.Middleware, controller controllers.Controller, journal controllers.Journal, group *gin.RouterGroup) Handler {
	h := &handler{Middleware: middleware, controller: controller, journalController: journal, Group: group}

	generate := group.Group("/generate", h.MemberAuth())
	{
		generate.POST("/marks", h.GenerateMarks)
		generate.POST("/absences", h.GenerateAbsences)
	}

	group.GET("/lessons/:id", h.Auth(), h.GetLesson)

	group.GET("/options", h.MemberAuth(), h.GetJournalAvailableOptions)
	group.GET("/:group/:subject/:teacher", h.MemberAuth(), h.GetJournal)
	group.GET("", h.MemberAuth(), h.GetUserJournal)

	mark := group.Group("/marks", h.MemberAuth("editJournal"))
	{
		mark.POST("list", h.AddMarks)
		mark.POST("", h.AddMark)
		mark.PUT("", h.UpdateMark)
		mark.DELETE(":id", h.DeleteMark)
	}

	absences := group.Group("/absences", h.MemberAuth("editJournal"))
	{
		absences.POST("list", h.AddAbsences)
		absences.POST("", h.AddAbsence)
		absences.PUT("", h.UpdateAbsence)
		absences.DELETE(":id", h.DeleteAbsence)
	}

	return h
}

// GenerateMarks godoc
// @Router /generate/marks [post]
func (j *handler) GenerateMarks(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var config dtos.MarksReport
	if err := ctx.BindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	file, err := j.controller.GenerateMarksReport(ctx, config, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	_, _ = file.WriteTo(ctx.Writer)
}

// GenerateAbsences godoc
// @Router /generate/absences [post]
func (j *handler) GenerateAbsences(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var config dtos.AbsencesReport
	if err := ctx.BindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	file, err := j.controller.GenerateAbsencesReport(ctx, config, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	_, _ = file.WriteTo(ctx.Writer)
}

// GetJournalAvailableOptions godoc
// @Router /options [get]
func (j *handler) GetJournalAvailableOptions(ctx *gin.Context) {
	user := j.GetUser(ctx)

	options, err := j.journalController.GetAvailableOptions(ctx, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, options)
}

// GetJournal godoc
// @Param group path string true "Group"
// @Param subject path string true "Subject"
// @Param teacher path string true "Teacher"
// @Router /{group}/{subject}/{teacher} [get]
func (j *handler) GetJournal(ctx *gin.Context) {
	user := j.GetUser(ctx)

	group := ctx.Param("groupID")
	subject := ctx.Param("subjectID")
	teacher := ctx.Param("teacherID")

	journal, err := j.journalController.BuildSubjectsJournal(ctx, group, subject, teacher, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

// GetUserJournal godoc
// @Router / [get]
func (j *handler) GetUserJournal(ctx *gin.Context) {
	user := j.GetUser(ctx)

	group := ctx.Query("groupID")
	subject := ctx.Query("subjectID")
	teacher := ctx.Query("teacherID")

	var journal entities.Journal
	var err error
	if group == "" || subject == "" || teacher == "" {
		journal, err = j.journalController.BuildStudentsJournal(ctx, user)
	} else {
		journal, err = j.journalController.BuildSubjectsJournal(ctx, group, subject, teacher, user)
	}

	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

// AddMarks godoc
// @Router /mark/list [post]
func (j *handler) AddMarks(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var marksDTO []dtos.AddMarkDTO
	if err := ctx.BindJSON(&marksDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.AddMarks(ctx, marksDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// AddMark godoc
// @Router /mark [post]
func (j *handler) AddMark(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var markDTO dtos.AddMarkDTO
	if err := ctx.BindJSON(&markDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.AddMark(ctx, markDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// UpdateMark godoc
// @Router /mark [put]
func (j *handler) UpdateMark(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var markDTO dtos.UpdateMarkDTO
	if err := ctx.BindJSON(&markDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.UpdateMark(ctx, user, markDTO)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// DeleteMark godoc
// @Param id path string true "Mark ID"
// @Router /mark/{id} [delete]
func (j *handler) DeleteMark(ctx *gin.Context) {
	user := j.GetUser(ctx)

	markId := ctx.Param("id")

	lesson, err := j.controller.DeleteMark(ctx, user, markId)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// AddAbsences godoc
// @Router /absences/list [post]
func (j *handler) AddAbsences(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var absencesDTO []dtos.AddAbsencesDTO
	if err := ctx.BindJSON(&absencesDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	absences, err := j.controller.AddAbsences(ctx, absencesDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, absences)
}

// AddAbsence godoc
// @Router /absences [post]
func (j *handler) AddAbsence(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var absencesDTO dtos.AddAbsencesDTO
	if err := ctx.BindJSON(&absencesDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.AddAbsence(ctx, absencesDTO, user)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// UpdateAbsence godoc
// @Router /absences [put]
func (j *handler) UpdateAbsence(ctx *gin.Context) {
	user := j.GetUser(ctx)

	var absences dtos.UpdateAbsencesDTO
	if err := ctx.BindJSON(&absences); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.UpdateAbsence(ctx, user, absences)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// DeleteAbsence godoc
// @Param id path string true "Absence ID"
// @Router /absences/{id} [delete]
func (j *handler) DeleteAbsence(ctx *gin.Context) {
	user := j.GetUser(ctx)

	absencesID := ctx.Param("id")
	lesson, err := j.controller.DeleteAbsence(ctx, user, absencesID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

// GetLesson godoc
// @Param id path string true "Lesson ID"
// @Router /lessons/{id} [delete]
func (j *handler) GetLesson(ctx *gin.Context) {
	user := j.GetUser(ctx)

	id := ctx.Param("id")
	studentID := ctx.Query("studentID")

	lessonInfo, err := j.controller.GetLessonInfo(ctx, user, studentID, id)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, lessonInfo)
}

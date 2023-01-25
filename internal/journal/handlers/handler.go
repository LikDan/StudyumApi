package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/global"
	"studyum/internal/journal/controllers"
	"studyum/internal/journal/dtos"
)

type Handler interface {
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

	GenerateMarks(ctx *gin.Context)
	GenerateAbsences(ctx *gin.Context)
}

type handler struct {
	global.Handler

	controller        controllers.Controller
	journalController controllers.Journal

	Group *gin.RouterGroup
}

func NewJournalHandler(authHandler global.Handler, controller controllers.Controller, journal controllers.Journal, group *gin.RouterGroup) Handler {
	h := &handler{Handler: authHandler, controller: controller, journalController: journal, Group: group}

	group.GET("/options", h.Auth(), h.GetJournalAvailableOptions)
	group.GET("/:group/:subject/:teacher", h.Auth(), h.GetJournal)
	group.GET("", h.Auth(), h.GetUserJournal)

	mark := group.Group("/mark", h.Auth("editJournal"))
	{
		mark.POST("list", h.AddMarks)
		mark.POST("", h.AddMark)
		mark.PUT("", h.UpdateMark)
		mark.DELETE(":id", h.DeleteMark)
	}

	absences := group.Group("/absences", h.Auth("editJournal"))
	{
		absences.POST("list", h.AddAbsences)
		absences.POST("", h.AddAbsence)
		absences.PUT("", h.UpdateAbsence)
		absences.DELETE(":id", h.DeleteAbsence)
	}

	generate := group.Group("/generate", h.Auth())
	{
		generate.POST("/marks", h.Auth(), h.GenerateMarks)
		generate.POST("/absences", h.Auth(), h.GenerateAbsences)
	}

	return h
}

func (j *handler) GenerateMarks(ctx *gin.Context) {
	user := j.Handler.GetUserViaCtx(ctx)

	var config dtos.MarksReport
	if err := ctx.BindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	file, err := j.controller.GenerateMarksReport(ctx, config, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	file.WriteTo(ctx.Writer)
}

func (j *handler) GenerateAbsences(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var config dtos.AbsencesReport
	if err := ctx.BindJSON(&config); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	file, err := j.controller.GenerateAbsencesReport(ctx, config, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	file.WriteTo(ctx.Writer)
}

func (j *handler) GetJournalAvailableOptions(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	options, err := j.journalController.BuildAvailableOptions(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, options)
}

func (j *handler) GetJournal(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	group := ctx.Param("group")
	subject := ctx.Param("subject")
	teacher := ctx.Param("teacher")

	journal, err := j.journalController.BuildSubjectsJournal(ctx, group, subject, teacher, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *handler) GetUserJournal(ctx *gin.Context) {
	user := j.Handler.GetUserViaCtx(ctx)

	journal, err := j.journalController.BuildStudentsJournal(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *handler) AddMarks(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var marks []dtos.AddMarkDTO
	if err := ctx.BindJSON(&marks); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.AddMarks(ctx, marks, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (j *handler) AddMark(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var mark dtos.AddMarkDTO
	if err := ctx.BindJSON(&mark); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := j.controller.AddMark(ctx, mark, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (j *handler) GetMark(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")

	lessons, err := j.controller.GetMark(ctx, group, subject, userIdHex, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (j *handler) UpdateMark(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var mark dtos.UpdateMarkDTO
	if err := ctx.BindJSON(&mark); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err := j.controller.UpdateMark(ctx, user, mark)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, mark)
}

func (j *handler) DeleteMark(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	markId := ctx.Param("id")

	err := j.controller.DeleteMark(ctx, user, markId)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, markId)
}

func (j *handler) AddAbsences(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var absencesDTO []dtos.AddAbsencesDTO
	if err := ctx.BindJSON(&absencesDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	absences, err := j.controller.AddAbsences(ctx, absencesDTO, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, absences)
}

func (j *handler) AddAbsence(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var absencesDTO dtos.AddAbsencesDTO
	if err := ctx.BindJSON(&absencesDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	absences, err := j.controller.AddAbsence(ctx, absencesDTO, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, absences)
}

func (j *handler) UpdateAbsence(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	var absences dtos.UpdateAbsencesDTO
	if err := ctx.BindJSON(&absences); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := j.controller.UpdateAbsence(ctx, user, absences); err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, absences)
}

func (j *handler) DeleteAbsence(ctx *gin.Context) {
	user := j.GetUserViaCtx(ctx)

	absencesID := ctx.Param("id")
	if err := j.controller.DeleteAbsence(ctx, user, absencesID); err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, absencesID)
}

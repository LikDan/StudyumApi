package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/dto"
	"studyum/internal/utils"
)

type JournalHandler interface {
	GetJournalAvailableOptions(ctx *gin.Context)

	GetJournal(ctx *gin.Context)
	GetUserJournal(ctx *gin.Context)

	AddMark(ctx *gin.Context)
	GetMark(ctx *gin.Context)
	UpdateMark(ctx *gin.Context)
	DeleteMark(ctx *gin.Context)
}

type journalHandler struct {
	Handler

	controller controllers.JournalController

	Group *gin.RouterGroup
}

func NewJournalHandler(authHandler Handler, controller controllers.JournalController, group *gin.RouterGroup) JournalHandler {
	h := &journalHandler{Handler: authHandler, controller: controller, Group: group}

	group.GET("/options", h.Auth(), h.GetJournalAvailableOptions)
	group.GET("/:group/:subject/:teacher", h.Auth(), h.GetJournal)
	group.GET("", h.Auth(), h.GetUserJournal)

	mark := group.Group("/mark", h.Auth("editJournal"))
	{
		mark.POST("", h.AddMark)
		mark.GET("", h.GetMark)
		mark.PUT("", h.UpdateMark)
		mark.DELETE("", h.DeleteMark)
	}
	return h
}

func (j *journalHandler) GetJournalAvailableOptions(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	options, err := j.controller.GetJournalAvailableOptions(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, options)
}

func (j *journalHandler) GetJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	group := ctx.Param("group")
	subject := ctx.Param("subject")
	teacher := ctx.Param("teacher")

	journal, err := j.controller.GetJournal(ctx, group, subject, teacher, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *journalHandler) GetUserJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	journal, err := j.controller.GetUserJournal(ctx, user)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (j *journalHandler) AddMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark dto.AddMarkDTO
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

func (j *journalHandler) GetMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

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

func (j *journalHandler) UpdateMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark dto.UpdateMarkDTO
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

func (j *journalHandler) DeleteMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	markId := ctx.Query("id")

	err := j.controller.DeleteMark(ctx, user, markId)
	if err != nil {
		j.Error(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, markId)
}

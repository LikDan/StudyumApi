package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"studyum/src/controllers"
	"studyum/src/models"
	"studyum/src/utils"
)

type JournalHandler struct {
	IAuthHandler

	controller controllers.IJournalController

	Group *gin.RouterGroup
}

func NewJournalHandler(authHandler IAuthHandler, controller controllers.IJournalController, group *gin.RouterGroup) *JournalHandler {
	h := &JournalHandler{IAuthHandler: authHandler, controller: controller, Group: group}

	group.GET("/options", h.Auth(), h.GetJournalAvailableOptions)
	group.GET("/:group/:subject/:teacher", h.Auth(), h.GetJournal)
	group.GET("", h.Auth(), h.GetUserJournal)

	mark := group.Group("/mark", h.Auth())
	{
		mark.POST("", h.AddMark)
		mark.GET("", h.GetMark)
		mark.PUT("", h.UpdateMark)
		mark.DELETE("", h.DeleteMark)
	}
	return h
}

func (h *JournalHandler) GetJournalAvailableOptions(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	options, err := h.controller.GetJournalAvailableOptions(ctx, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, options)
}

func (h *JournalHandler) GetJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	group := ctx.Param("group")
	subject := ctx.Param("subject")
	teacher := ctx.Param("teacher")

	journal, err := h.controller.GetJournal(ctx, group, subject, teacher, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (h *JournalHandler) GetUserJournal(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	journal, err := h.controller.GetUserJournal(ctx, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, journal)
}

func (h *JournalHandler) AddMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark models.Mark
	if err := ctx.BindJSON(&mark); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	lesson, err := h.controller.AddMark(ctx, mark, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (h *JournalHandler) GetMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	group := ctx.Query("group")
	subject := ctx.Query("subject")
	userIdHex := ctx.Query("userId")

	lessons, err := h.controller.GetMark(ctx, group, subject, userIdHex, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, lessons)
}

func (h *JournalHandler) UpdateMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	var mark models.Mark
	if err := ctx.BindJSON(&mark); models.BindError(err, 400, models.UNDEFINED).CheckAndResponse(ctx) {
		return
	}

	lesson, err := h.controller.UpdateMark(ctx, mark, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

func (h *JournalHandler) DeleteMark(ctx *gin.Context) {
	user := utils.GetUserViaCtx(ctx)

	markId := ctx.Query("markId")
	userId := ctx.Query("userId")
	subjectId := ctx.Query("subjectId")

	lesson, err := h.controller.DeleteMark(ctx, markId, userId, subjectId, user)
	if err.CheckAndResponse(ctx) {
		return
	}

	ctx.JSON(http.StatusOK, lesson)
}

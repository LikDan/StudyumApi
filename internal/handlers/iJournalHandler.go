package handlers

import "github.com/gin-gonic/gin"

type IJournalHandler interface {
	GetJournalAvailableOptions(ctx *gin.Context)

	GetJournal(ctx *gin.Context)
	GetUserJournal(ctx *gin.Context)

	AddMark(ctx *gin.Context)
	GetMark(ctx *gin.Context)
	UpdateMark(ctx *gin.Context)
	DeleteMark(ctx *gin.Context)
}

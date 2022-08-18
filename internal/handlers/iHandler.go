package handlers

import "github.com/gin-gonic/gin"

type IHandler interface {
	Auth(permissions ...string) gin.HandlerFunc
	Error(ctx *gin.Context, err error)
}

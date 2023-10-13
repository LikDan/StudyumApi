package main

import "github.com/gin-gonic/gin"

type Api struct {
	Default *gin.RouterGroup
	V1      *gin.RouterGroup
}

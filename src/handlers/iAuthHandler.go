package handlers

import "github.com/gin-gonic/gin"

type IAuthHandler interface {
	Auth() gin.HandlerFunc
}

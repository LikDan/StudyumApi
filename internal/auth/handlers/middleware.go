package handlers

import (
	"github.com/gin-gonic/gin"
	"studyum/internal/auth/controllers"
	"studyum/internal/global"
	"studyum/pkg/jwt"
)

type Middleware interface {
	Auth() gin.HandlerFunc
	TryAuth() gin.HandlerFunc
	MemberAuth(permissions ...string) gin.HandlerFunc

	SetTokenPairCookie(ctx *gin.Context, pair jwt.TokenPair)
}

type middleware struct {
	global.Handler

	controller controllers.Middleware
}

func NewMiddleware(handler global.Handler, controller controllers.Middleware) Middleware {
	return &middleware{Handler: handler, controller: controller}
}

func (h *middleware) SetTokenPairCookie(ctx *gin.Context, pair jwt.TokenPair) {
	ctx.SetCookie("refresh", pair.Refresh, 60*60*24*30, "/", "", false, true)
	ctx.SetCookie("access", pair.Access, 60*15, "/", "", false, true)
}

func (h *middleware) tokenPair(ctx *gin.Context) jwt.TokenPair {
	refresh := ctx.GetString("refresh") //proceed on oauth2
	if refresh == "" {
		refresh, _ = ctx.Cookie("refresh")
	}

	access, _ := ctx.Cookie("access")
	return jwt.TokenPair{Access: access, Refresh: refresh}
}

func (h *middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pair := h.tokenPair(ctx)
		newPair, update, user, err := h.controller.Auth(ctx, pair, ctx.ClientIP())
		if err != nil {
			h.Error(ctx, err)
			return
		}
		if update {
			h.SetTokenPairCookie(ctx, newPair)
		}

		ctx.Set("user", user)

		ctx.Next()
		if ctx.Request.Context().Err() != nil && update {
			if err = h.controller.Recover(ctx, pair, newPair, ctx.ClientIP(), user.Id); err != nil {
				h.Error(ctx, err)
				return
			}
		}
	}
}

func (h *middleware) TryAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pair := h.tokenPair(ctx)
		newPair, update, user, err := h.controller.Auth(ctx, pair, ctx.ClientIP())
		if err != nil {
			return
		}
		if update {
			h.SetTokenPairCookie(ctx, newPair)
		}

		ctx.Set("user", user)

		ctx.Next()
		if ctx.Request.Context().Err() != nil && update {
			if err = h.controller.Recover(ctx, pair, newPair, ctx.ClientIP(), user.Id); err != nil {
				h.Error(ctx, err)
				return
			}
		}
	}
}

func (h *middleware) MemberAuth(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pair := h.tokenPair(ctx)
		newPair, update, user, err := h.controller.Auth(ctx, pair, ctx.ClientIP(), permissions...)
		if err != nil {
			h.Error(ctx, err)
			return
		}
		if update {
			h.SetTokenPairCookie(ctx, newPair)
		}

		ctx.Set("user", user)

		ctx.Next()
		if ctx.Request.Context().Err() != nil && update {
			if err = h.controller.Recover(ctx, pair, newPair, ctx.ClientIP(), user.Id); err != nil {
				h.Error(ctx, err)
				return
			}
		}
	}
}

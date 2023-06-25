package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	"studyum/internal/utils"
	entities2 "studyum/pkg/jwt/entities"
)

type Middleware interface {
	GrpcAuth(ctx context.Context, pair entities2.TokenPair) (entities2.TokenPair, bool, entities.User, error)

	Auth() gin.HandlerFunc
	TryAuth() gin.HandlerFunc
	MemberAuth(permissions ...string) gin.HandlerFunc

	SetTokenPairCookie(ctx *gin.Context, pair entities2.TokenPair)
	SetTokenPairHeader(ctx *gin.Context, pair entities2.TokenPair)
	DeleteTokenPairCookie(ctx *gin.Context)

	GetUser(ctx *gin.Context) entities.User
}

type middleware struct {
	controller controllers.Middleware
}

func NewMiddleware(controller controllers.Middleware) Middleware {
	return &middleware{controller: controller}
}

func (h *middleware) SetTokenPairCookie(ctx *gin.Context, pair entities2.TokenPair) {
	ctx.SetCookie("refresh", pair.Refresh, 60*60*24*30, "/", "", true, true)
	ctx.SetCookie("access", pair.Access, 60*15, "/", "", true, true)
}

func (h *middleware) SetTokenPairHeader(ctx *gin.Context, pair entities2.TokenPair) {
	ctx.Header("SetAccessToken", pair.Access)
	ctx.Header("SetRefreshToken", pair.Refresh)
}

func (h *middleware) DeleteTokenPairCookie(ctx *gin.Context) {
	ctx.SetCookie("refresh", "", 0, "/", "", true, true)
	ctx.SetCookie("access", "", 0, "/", "", true, true)
}

func (h *middleware) tokenPair(ctx *gin.Context) entities2.TokenPair {
	access := ctx.GetHeader("Authorization")
	if access != "" {
		return entities2.TokenPair{
			Access:  access,
			Refresh: "",
		}
	}

	refresh := ctx.GetString("refresh") //proceed on oauth2
	if refresh == "" {
		refresh, _ = ctx.Cookie("refresh")
	}

	access, _ = ctx.Cookie("access")
	return entities2.TokenPair{Access: access, Refresh: refresh}
}

func (h *middleware) authViaApiToken(ctx *gin.Context) bool {
	apiToken := ctx.GetHeader("ApiToken")
	if apiToken == "" {
		return false
	}

	user, err := h.controller.AuthViaApiToken(ctx, apiToken)
	if err != nil {
		_ = ctx.Error(err)
		ctx.Abort()
		return true
	}

	ctx.Set("user", user)
	return true
}

func (h *middleware) GrpcAuth(ctx context.Context, pair entities2.TokenPair) (entities2.TokenPair, bool, entities.User, error) {
	return h.controller.Auth(ctx, pair, "")
}

func (h *middleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.authViaApiToken(ctx) {
			return
		}

		pair := h.tokenPair(ctx)
		newPair, update, user, err := h.controller.Auth(ctx, pair, ctx.ClientIP())
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
		if update {
			h.SetTokenPairCookie(ctx, newPair)
		}

		ctx.Set("user", user)

		ctx.Next()
	}
}

func (h *middleware) TryAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.authViaApiToken(ctx) {
			return
		}

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
		//if ctx.Request.Context().Err() != nil && update {
		//	if err = h.controller.Recover(ctx, pair, newPair, ctx.ClientIP(), user.Id); err != nil {
		//		_ = ctx.Error(err)
		//		ctx.Abort()
		//		return
		//	}
		//}
	}
}

func (h *middleware) MemberAuth(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if h.authViaApiToken(ctx) {
			return
		}

		pair := h.tokenPair(ctx)
		newPair, update, user, err := h.controller.Auth(ctx, pair, ctx.ClientIP(), permissions...)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
		if update {
			h.SetTokenPairCookie(ctx, newPair)
		}

		ctx.Set("user", user)

		ctx.Next()
		//if ctx.Request.Context().Err() != nil && update {
		//	if err = h.controller.Recover(ctx, pair, newPair, ctx.ClientIP(), user.Id); err != nil {
		//		_ = ctx.Error(err)
		//		ctx.Abort()
		//		return
		//	}
		//}
	}
}

func (h *middleware) GetUser(ctx *gin.Context) entities.User {
	return utils.GetViaCtx[entities.User](ctx, "user")
}

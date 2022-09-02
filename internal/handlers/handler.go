package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"studyum/internal/controllers"
)

type Handler interface {
	Auth(permissions ...string) gin.HandlerFunc
	AuthToken(permissions ...string) gin.HandlerFunc
	Error(ctx *gin.Context, err error)
}

type handler struct {
	controller controllers.Controller
}

func NewHandler(controller controllers.Controller) Handler {
	return &handler{controller: controller}
}

func (h *handler) AuthToken(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("authToken")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, "no token")
			ctx.Abort()
			return
		}

		user, err := h.controller.Auth(ctx, token, permissions...)
		if err != nil {
			h.Error(ctx, err)
			return
		}

		ctx.Set("user", user)
	}
}

func (h *handler) Auth(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authentication")
		if token == "" {
			h.AuthToken(permissions...)(ctx)
			return
		}

		user, err := h.controller.AuthJWT(ctx, token, permissions...)
		if err != nil {
			h.Error(ctx, err)
			return
		}

		ctx.Set("user", user)
	}
}

func (h *handler) Error(ctx *gin.Context, err error) {
	var code int

	switch {
	case
		errors.Is(mongo.ErrMissingResumeToken, err):
		code = http.StatusBadGateway
		break
	case
		errors.Is(mongo.ErrUnacknowledgedWrite, err),
		errors.Is(mongo.ErrClientDisconnected, err):
		code = http.StatusInternalServerError
		break
	case
		errors.Is(mongo.ErrNilDocument, err),
		errors.Is(mongo.ErrNoDocuments, err),
		errors.Is(mongo.ErrNilValue, err),
		errors.Is(mongo.ErrEmptySlice, err),
		errors.Is(mongo.ErrNilCursor, err),
		errors.Is(controllers.NotValidParams, err):
		code = http.StatusBadRequest
		break
	case errors.Is(controllers.NotAuthorizationError, err):
		code = http.StatusUnauthorized
		break
	case errors.Is(controllers.NoPermission, err):
		code = http.StatusForbidden
		break
	default:
		code = http.StatusInternalServerError
	}

	ctx.JSON(code, err.Error())
	_ = ctx.Error(err)
	ctx.Abort()
}

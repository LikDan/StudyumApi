package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/repositories"
)

type Handler struct {
	controller controllers.IController
}

func NewHandler(controller controllers.IController) *Handler {
	return &Handler{controller: controller}
}

func (h *Handler) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("authToken")
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, "no token")
			ctx.Abort()
			return
		}

		user, err := h.controller.Auth(ctx, token)
		if err != nil {
			h.Error(ctx, err)
			return
		}

		ctx.Set("user", user)
	}
}

func (h *Handler) Error(ctx *gin.Context, err error) {
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
		errors.Is(mongo.ErrNilValue, err),
		errors.Is(mongo.ErrEmptySlice, err),
		errors.Is(mongo.ErrNilCursor, err),
		errors.Is(controllers.NotValidParams, err):
		code = http.StatusBadRequest
		break
	case errors.Is(repositories.NotAuthorizationError, err):
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

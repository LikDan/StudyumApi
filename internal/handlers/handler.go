package handlers

import (
	"github.com/gin-gonic/gin"
	j "github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"studyum/internal/controllers"
	"studyum/internal/controllers/validators"
	"studyum/pkg/jwt"
)

type Handler interface {
	User(permissions ...string) gin.HandlerFunc
	Auth(permissions ...string) gin.HandlerFunc

	Error(ctx *gin.Context, err error)

	SetTokenPairCookie(ctx *gin.Context, pair jwt.TokenPair)
}

type handler struct {
	controller controllers.Controller
}

func NewHandler(controller controllers.Controller) Handler {
	return &handler{controller: controller}
}

func (h *handler) authViaAccessToken(ctx *gin.Context, permissions ...string) error {
	token, err := ctx.Cookie("access")
	if err != nil {
		return err
	}

	user, err := h.controller.AuthJWT(ctx, token, permissions...)
	if err != nil {
		return err
	}

	ctx.Set("user", user)
	return nil
}

func (h *handler) authViaRefreshToken(ctx *gin.Context, permissions ...string) error {
	refreshToken, err := ctx.Cookie("refresh")
	if err != nil {
		return err
	}

	user, pair, err := h.controller.AuthJWTByRefreshToken(ctx, refreshToken, permissions...)
	if err != nil {
		return err
	}

	h.SetTokenPairCookie(ctx, pair)

	ctx.Set("user", user)
	return nil
}

func (h *handler) auth(ctx *gin.Context, permissions ...string) error {
	if err := h.authViaAccessToken(ctx, permissions...); err != nil {
		if err = h.authViaRefreshToken(ctx, permissions...); err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) Auth(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := h.auth(ctx, permissions...)
		if err != nil {
			h.Error(ctx, err)
		}
	}
}

func (h *handler) SetTokenPairCookie(ctx *gin.Context, pair jwt.TokenPair) {
	ctx.SetCookie("refresh", pair.Refresh, 60*60*24*30, "/", "", false, true)
	ctx.SetCookie("access", pair.Access, 60*15, "/", "", false, true)
}

func (h *handler) User(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_ = h.auth(ctx, permissions...)
	}
}

func (h *handler) Error(ctx *gin.Context, err error) {
	var code int

	switch {
	case
		errors.Is(err, mongo.ErrMissingResumeToken):
		code = http.StatusBadGateway
		break
	case
		errors.Is(err, mongo.ErrUnacknowledgedWrite),
		errors.Is(err, mongo.ErrClientDisconnected):
		code = http.StatusInternalServerError
		break
	case
		errors.Is(err, mongo.ErrNilDocument),
		errors.Is(err, mongo.ErrNoDocuments),
		errors.Is(err, mongo.ErrNilValue),
		errors.Is(err, mongo.ErrEmptySlice),
		errors.Is(err, mongo.ErrNilCursor),
		errors.Is(err, controllers.NotValidParams),
		errors.Is(err, validators.ValidationError):
		code = http.StatusBadRequest
		break
	case errors.Is(err, controllers.NotAuthorizationError),
		errors.Is(err, j.ErrSignatureInvalid):
		code = http.StatusUnauthorized
		break
	case errors.Is(err, controllers.NoPermission):
		code = http.StatusForbidden
		break
	default:
		code = http.StatusInternalServerError
	}

	ctx.JSON(code, err.Error())
	_ = ctx.Error(err)
	ctx.Abort()
}

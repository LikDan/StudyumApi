package middlewares

import (
	"github.com/gin-gonic/gin"
	j "github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	auth "studyum/internal/auth/controllers"
	"studyum/internal/journal/controllers"
	"studyum/internal/schedule"
	"studyum/pkg/datetime"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		if len(ctx.Errors) == 0 {
			return
		}

		code := GetHttpCodeByError(ctx.Errors[0])
		ctx.JSON(code, ctx.Errors[0].Error())
	}
}

func GetHttpCodeByError(err error) int {
	var code int

	switch {
	case
		errors.Is(err, mongo.ErrMissingResumeToken):
		code = http.StatusBadGateway
	case
		errors.Is(err, mongo.ErrUnacknowledgedWrite),
		errors.Is(err, mongo.ErrClientDisconnected):
		code = http.StatusInternalServerError
	case
		errors.Is(err, mongo.ErrNilDocument),
		errors.Is(err, mongo.ErrNoDocuments),
		errors.Is(err, mongo.ErrNilValue),
		errors.Is(err, mongo.ErrEmptySlice),
		errors.Is(err, mongo.ErrNilCursor),
		errors.Is(err, auth.ValidationError),
		errors.Is(err, auth.BadClaimsErr),
		errors.Is(err, auth.ErrExpired),
		errors.Is(err, datetime.DurationError),
		errors.Is(err, controllers.NotValidParams),
		errors.Is(err, schedule.NotValidParams),
		errors.Is(err, schedule.ValidationError):
		code = http.StatusBadRequest
	case errors.Is(err, j.ErrSignatureInvalid),
		errors.Is(err, http.ErrNoCookie):
		code = http.StatusUnauthorized
	case errors.Is(err, auth.ForbiddenErr),
		errors.Is(err, controllers.ErrNoPermission):
		code = http.StatusForbidden
	default:
		code = http.StatusInternalServerError
	}

	return code
}

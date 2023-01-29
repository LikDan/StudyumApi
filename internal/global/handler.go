package global

import (
	"github.com/gin-gonic/gin"
	j "github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	auth "studyum/internal/auth/controllers"
	"studyum/internal/auth/entities"
	"studyum/internal/utils"
	"studyum/pkg/datetime"
)

type Handler interface {
	GetUser(ctx *gin.Context) entities.User
	Error(ctx *gin.Context, err error)
}

type handler struct {
	controller Controller
}

func NewHandler(controller Controller) Handler {
	return &handler{controller: controller}
}

func (h *handler) GetUser(ctx *gin.Context) entities.User {
	return utils.GetViaCtx[entities.User](ctx, "user")
}

func (h *handler) Error(ctx *gin.Context, err error) {
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
		errors.Is(err, NotValidParams),
		errors.Is(err, ValidationError),
		errors.Is(err, auth.ValidationError),
		errors.Is(err, datetime.DurationError),
		errors.Is(err, auth.BadClaimsErr):
		code = http.StatusBadRequest
	case errors.Is(err, NotAuthorizationError),
		errors.Is(err, j.ErrSignatureInvalid),
		errors.Is(err, http.ErrNoCookie):
		code = http.StatusUnauthorized
	case errors.Is(err, NoPermission),
		errors.Is(err, ForbiddenError),
		errors.Is(err, auth.ForbiddenErr):
		code = http.StatusForbidden
	default:
		code = http.StatusInternalServerError
	}

	ctx.JSON(code, err.Error())
	_ = ctx.Error(err)
	ctx.Abort()
}

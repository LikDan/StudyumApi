package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"studyum/grpc/auth/protoauth"
	"studyum/internal/auth/controllers"
	"studyum/internal/auth/dto"
	"studyum/pkg/jwt/entities"
)

type Auth struct {
	Middleware

	controller controllers.Auth

	Group *gin.RouterGroup
}

func NewAuth(middleware Middleware, controller controllers.Auth, group *gin.RouterGroup, grpcServer *grpc.Server) *Auth {
	h := &Auth{Middleware: middleware, controller: controller, Group: group}

	protoauth.RegisterAuthServer(grpcServer, h)

	group.PUT("updateToken", h.UpdateByRefreshToken)

	group.PUT("login", h.Login)

	group.POST("signup", h.SignUp)
	group.PUT("signup/stage1", h.Auth(), h.SignUpUserStage1)
	group.POST("signup/code", h.Auth(), h.SignUpStage1ViaCode)
	group.DELETE("signout", h.Auth(), h.SignOut)

	group.POST("email/confirm", h.Auth(), h.ConfirmEmail)
	group.POST("email/resendCode", h.Auth(), h.ResendEmailCode)

	group.DELETE("sessions", h.Auth(), h.TerminateAllSessions)

	return h
}

// UpdateByRefreshToken godoc
// @Router /updateToken [put]
func (h *Auth) UpdateByRefreshToken(ctx *gin.Context) {
	var token string
	if err := ctx.BindJSON(&token); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	pair, err := h.controller.UpdateByRefreshToken(ctx, token, ctx.ClientIP())
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairHeader(ctx, pair)
	ctx.Status(http.StatusOK)
}

// Login godoc
// @Router /login [put]
func (h *Auth) Login(ctx *gin.Context) {
	var data dto.Login
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := h.controller.Login(ctx, ctx.ClientIP(), data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)
	h.SetTokenPairHeader(ctx, pair)

	ctx.JSON(http.StatusOK, user)
}

// SignUp godoc
// @Router /signup [post]
func (h *Auth) SignUp(ctx *gin.Context) {
	var data dto.SignUp
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, pair, err := h.controller.SignUp(ctx, ctx.ClientIP(), data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.SetTokenPairCookie(ctx, pair)
	h.SetTokenPairHeader(ctx, pair)
	ctx.JSON(http.StatusOK, user)
}

// SignUpUserStage1 godoc
// @Router /signup/stage1 [put]
func (h *Auth) SignUpUserStage1(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.SignUpStage1
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.controller.SignUpStage1(ctx, user, data)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// SignUpStage1ViaCode godoc
// @Router /signup/code [post]
func (h *Auth) SignUpStage1ViaCode(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.SignUpWithCode
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.controller.SignUpStage1ViaCode(ctx, user, data.Code)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// SignOut godoc
// @Router /signout [delete]
func (h *Auth) SignOut(ctx *gin.Context) {
	token, _ := ctx.Cookie("refresh")
	if token != "" {
		if err := h.controller.SignOut(ctx, token); err != nil {
			if err != mongo.ErrNoDocuments {
				_ = ctx.Error(err)
				return
			}
		}
	}

	h.DeleteTokenPairCookie(ctx)
	ctx.Status(http.StatusNoContent)
}

// ConfirmEmail godoc
// @Router /email/confirm [post]
func (h *Auth) ConfirmEmail(ctx *gin.Context) {
	user := h.GetUser(ctx)

	var data dto.VerificationCode
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := h.controller.ConfirmEmail(ctx, user, data); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ResendEmailCode godoc
// @Router /email/resendCode [post]
func (h *Auth) ResendEmailCode(ctx *gin.Context) {
	user := h.GetUser(ctx)
	if err := h.controller.ResendEmailCode(ctx, user); err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// TerminateAllSessions godoc
// @Router /sessions [delete]
func (h *Auth) TerminateAllSessions(ctx *gin.Context) {
	user := h.GetUser(ctx)
	if err := h.controller.TerminateAll(ctx, user); err != nil {
		_ = ctx.Error(err)
		return
	}

	h.DeleteTokenPairCookie(ctx)
	ctx.Status(http.StatusNoContent)
}

func (h *Auth) AuthUser(ctx context.Context, request *protoauth.AuthRequest) (*protoauth.AuthResponse, error) {
	pair, update, user, err := h.GrpcAuth(ctx, entities.TokenPair{
		Access:  request.Jwt.Access,
		Refresh: request.Jwt.Refresh,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	successfully := true
	for _, requiredPermission := range request.RequiredPermissions {
		found := false
		for _, permission := range user.Permissions {
			if requiredPermission == permission {
				found = true
				break
			}
		}
		if !found {
			successfully = false
			break
		}
	}

	if !successfully {
		return nil, status.Errorf(codes.PermissionDenied, "no permission")
	}

	return &protoauth.AuthResponse{
		User: &protoauth.User{
			Id:           user.Id.Hex(),
			Name:         user.Name,
			StudyPlaceID: user.StudyPlaceID.Hex(),
		},
		Update: update,
		Jwt: &protoauth.JWT{
			Refresh: pair.Refresh,
			Access:  pair.Access,
		},
	}, nil
}

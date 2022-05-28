package models

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"studyum/src/api"
)

type Error struct {
	Error   error       `json:"error" bson:"error"`
	Code    int         `json:"code" bson:"code"`
	LogType api.LogType `json:"log" bson:"log"`
}

func (err Error) Check() bool {
	if err.Error == nil {
		return false
	}

	switch err.LogType {
	case api.UNDEFINED:
		break
	case api.INFO:
		logrus.Info(err.Error.Error())
		break
	case api.WARNING:
		logrus.Warning(err.Error.Error())
		break
	case api.ERROR:
		logrus.Error(err.Error.Error())
		break
	}
	return true
}

func (err Error) CheckAndResponse(ctx *gin.Context) bool {
	if !err.Check() {
		return false
	}

	ctx.JSON(err.Code, err.Error.Error())
	return true
}

func BindErrorStr(err string, code int, logType api.LogType) *Error {
	return BindError(errors.New(err), code, logType)
}

func BindError(err error, code int, logType api.LogType) *Error {
	return &Error{
		Error:   err,
		Code:    code,
		LogType: logType,
	}
}

func EmptyError() *Error {
	return &Error{
		Error:   nil,
		Code:    0,
		LogType: api.UNDEFINED,
	}
}

package kbp

import (
	"context"
	"github.com/pkg/errors"
	"studyum/internal/parser/appDTO"
)

func (a *app) GetSignUpDataByCode(context.Context, string) (appDTO.SignUpCode, error) {
	return appDTO.SignUpCode{}, errors.New("not implemented")
}

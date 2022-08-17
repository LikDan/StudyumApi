package handler

import (
	"studyum/internal/parser/entities"
)

type IHandler interface {
	Update(app entities.IApp)
}

package handler

import (
	"studyum/src/parser/entities"
)

type IHandler interface {
	Update(app entities.IApp)
}

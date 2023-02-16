package marks

import (
	"studyum/internal/apps/apps/kbp/shared"
	"studyum/internal/apps/entities"
	appShared "studyum/internal/apps/shared"
)

func New(shared appShared.Shared, auth shared.AuthRepository) entities.MarksManageInterface {
	repository := &repository{}

	return NewController(repository, shared, auth)
}

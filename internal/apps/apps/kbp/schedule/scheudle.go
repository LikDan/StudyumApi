package schedule

import (
	"go.mongodb.org/mongo-driver/mongo"
	"studyum/internal/apps/apps/kbp/shared"
	"studyum/internal/apps/entities"
	appShared "studyum/internal/apps/shared"
)

func New(shared appShared.Shared, db *mongo.Database, auth shared.AuthRepository) entities.LessonsManageInterface {
	r := NewRepository()
	m := NewMongoRepository(db)

	return NewController(r, m, shared, auth)
}

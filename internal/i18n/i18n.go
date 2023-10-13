package i18n

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"studyum/internal/i18n/controllers"
	"studyum/internal/i18n/handlers"
	"studyum/internal/i18n/repositories"
)

func New(session *sql.DB, group *gin.RouterGroup) (repositories.Repository, controllers.Controller, handlers.PublicHandler) {
	repository := repositories.NewI18nRepository(session)
	controller := controllers.NewController(repository)
	handler := handlers.NewPublicHandler(controller, group)
	return repository, controller, handler
}

package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"studyum/internal/i18n/controllers"
	"studyum/internal/i18n/handlers"
	"studyum/internal/i18n/repositories"
)

func New(session *pgxpool.Pool, group *gin.RouterGroup) (repositories.Repository, controllers.Controller, handlers.PublicHandler) {
	repository := repositories.NewI18nRepository(session)
	controller := controllers.NewController(repository)
	handler := handlers.NewPublicHandler(controller, group)
	return repository, controller, handler
}

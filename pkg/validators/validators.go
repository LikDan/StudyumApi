package validators

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"strings"
)

func required(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(string)
	return ok && len(strings.TrimSpace(date)) != 0
}

func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	_ = v.RegisterValidation("req", required)
}

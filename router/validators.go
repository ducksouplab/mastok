package router

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var namespaceValidator validator.Func = func(fl validator.FieldLevel) bool {
	alphaNumericRegexString := "^[a-zA-Z0-9_]+$"
	alphaNumericRegex := regexp.MustCompile(alphaNumericRegexString)
	return alphaNumericRegex.MatchString(fl.Field().String())
}

func addCustomValidators() {
	// add custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("namespace", namespaceValidator)
	}
}

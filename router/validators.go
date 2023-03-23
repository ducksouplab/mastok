package router

import (
	"regexp"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var alphaNumericRegex *regexp.Regexp

func init() {
	alphaNumericRegexString := "^[a-zA-Z0-9_]+$"
	alphaNumericRegex = regexp.MustCompile(alphaNumericRegexString)
}

var namespaceValidator validator.Func = func(fl validator.FieldLevel) bool {
	return alphaNumericRegex.MatchString(fl.Field().String())
}

var groupingValidator validator.Func = func(fl validator.FieldLevel) bool {
	_, ok := helpers.ParseGroup(fl.Field().String())
	return ok
}

func addCustomValidators() {
	// add custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("namespaceValidate", namespaceValidator)
		v.RegisterValidation("groupingValidate", groupingValidator)
	}
}

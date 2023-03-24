package router

import (
	"log"
	"regexp"
	"strings"

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
	grouping, ok := helpers.ParseGroup(fl.Field().String())
	if !ok {
		return false
	}
	// grouping is optional
	if len(grouping.Groups) == 0 {
		return true
	}
	// if grouping is present, we need to check the sum of groups is equal to PerSession
	var size int
	for _, g := range grouping.Groups {
		size += g.Size
	}
	perSessionValue, _, _, _ := fl.GetStructFieldOK2()
	return size == int(perSessionValue.Int())
}

func addCustomValidators() {
	// add custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("namespaceValidate", namespaceValidator)
		v.RegisterValidation("groupingValidate", groupingValidator)
	}
}

func changeErrorMessage(err string) string {
	log.Printf(">>>>>>>>>>> %v", err)
	if strings.Contains(err, "Campaign.Grouping") {
		return "Format invalid: grouping rule (or check the sum of groups matches 'Participants per session')"
	}
	output := strings.Replace(err, "UNIQUE constraint failed", "Already taken", 1)
	output = strings.Replace(output, "campaigns.", "", 1)
	return output
}

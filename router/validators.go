package router

import (
	"regexp"
	"strings"

	"github.com/ducksouplab/mastok/models"
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
	groupingString := fl.Field().String()
	// grouping is optional
	if len(groupingString) == 0 {
		return true
	}
	grouping, err := models.RawParseGroupingString(groupingString)
	if err != nil {
		return false
	}
	// if grouping is present, we need to check the sum of groups is equal to PerSession
	var size int
	for _, g := range grouping.Groups {
		size += g.Size
	}
	perSessionValue, _, _, _ := fl.GetStructFieldOK2()
	perSessionInt := int(perSessionValue.Int())
	return size == perSessionInt
}

func addCustomValidators() {
	// add custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("namespaceValidate", namespaceValidator)
		v.RegisterValidation("groupingValidate", groupingValidator)
	}
}

func changeErrorMessage(err string) string {
	if strings.Contains(err, "Campaign.Grouping") {
		return "Format invalid: grouping rule (or check the sum of groups matches 'Participants per session')"
	}
	output := strings.Replace(err, "UNIQUE constraint failed", "Already taken", 1)
	output = strings.Replace(output, "campaigns.", "", 1)
	return output
}

package router

import (
	"regexp"
	"strings"

	"github.com/ducksouplab/mastok/cache"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/models"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var alphaNumericRegex *regexp.Regexp

func init() {
	alphaNumericRegexString := "^[a-zA-Z0-9_]+$"
	alphaNumericRegex = regexp.MustCompile(alphaNumericRegexString)
}

var perSessionValidator validator.Func = func(fl validator.FieldLevel) bool {
	perSession := int(fl.Field().Int())
	oTreeConfigNameField, _, _, _ := fl.GetStructFieldOK2()
	oTreeConfigName := oTreeConfigNameField.String()
	config, err := cache.GetOTreeConfig(oTreeConfigName)
	if err != nil {
		// if no config is found, does not fail (makes test easier)
		return true
	}
	return helpers.Contains(config.NumParticipantsAllowed, perSession)
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
	perSessionField, _, _, _ := fl.GetStructFieldOK2()
	perSessionInt := int(perSessionField.Int())
	return size == perSessionInt
}

var consentValidator validator.Func = func(fl validator.FieldLevel) bool {
	consentString := fl.Field().String()
	ok := strings.Contains(consentString, "[accept]")
	ok = ok && strings.Contains(consentString, "[/accept]")
	return ok
}

func addCustomValidators() {
	// add custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("perSessionValidate", perSessionValidator)
		v.RegisterValidation("namespaceValidate", namespaceValidator)
		v.RegisterValidation("groupingValidate", groupingValidator)
		v.RegisterValidation("consentValidate", consentValidator)
	}
}

func changeErrorMessage(err string) string {
	if strings.Contains(err, "perSessionValidate") {
		return "Participants per session: check allowed values for this oTree experiment"
	}
	if strings.Contains(err, "Campaign.Grouping") {
		return "Format invalid: grouping rule (or check the sum of groups matches 'Participants per session')"
	}
	if strings.Contains(err, "Campaign.Consent") {
		return "Format invalid: consent needs an [accept]...[/accept] tag"
	}
	output := strings.Replace(err, "UNIQUE constraint failed", "Already taken", 1)
	output = strings.Replace(output, "campaigns.", "", 1)
	return output
}

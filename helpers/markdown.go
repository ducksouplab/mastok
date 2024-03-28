package helpers

import (
	"html/template"
	"strings"

	"github.com/shurcooL/github_flavored_markdown"
)

func MarkdownToHTML(markdown string) template.HTML {
	out := string(github_flavored_markdown.Markdown([]byte(markdown)))
	// custom formatting
	out = strings.Replace(out, "[accept]", "<button class=\"btn btn-success\">", 1)
	out = strings.Replace(out, "[/accept]", "</button>", 1)
	out = strings.Replace(out, "[alert]", "<div id=\"alert-container\" class=\"alert alert-danger\" role=\"alert\">", 1)
	out = strings.Replace(out, "[/alert]", "</div>", 1)
	out = strings.Replace(out, "[ducksoup_test]", "<a href=\"https://ducksoup.psy.gla.ac.uk/test/direct/\" target=\"_blank\">this link</a>", 1)
	out = strings.Replace(out, "[ducksoup_audio_test]", "<a href=\"https://ducksoup.psy.gla.ac.uk/test/audio_direct/\" target=\"_blank\">this link</a>", 1)

	return template.HTML(out)
}

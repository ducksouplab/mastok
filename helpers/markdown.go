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
	out = strings.Replace(out, "[alert]", "<div id=\"alert-container\" class=\"alert alert-warning\" role=\"alert\">", 1)
	out = strings.Replace(out, "[/alert]", "</div>", 1)
	return template.HTML(out)
}

package helpers

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/shurcooL/github_flavored_markdown"
)

func MarkdownToHTML(markdown string) template.HTML {
	re := regexp.MustCompile(`\[new_link\](.*?)\[end_link\]`)
	markdown = re.ReplaceAllString(markdown, `<a href="$1" target="_blank">this link</a>`)

	fmt.Println("In between markdown:", markdown) // Debug output

	out := string(github_flavored_markdown.Markdown([]byte(markdown)))
	// custom formatting
	out = strings.Replace(out, "[accept]", "<button class=\"btn btn-success\">", 1)
	out = strings.Replace(out, "[/accept]", "</button>", 1)
	out = strings.Replace(out, "[alert]", "<div id=\"alert-container\" class=\"alert alert-danger\" role=\"alert\">", 1)
	out = strings.Replace(out, "[/alert]", "</div>", 1)
	out = strings.Replace(out, "[ducksoup_test]", "<a href=\"https://ducksoup.psy.gla.ac.uk/test/direct/\" target=\"_blank\">this link</a>", 1)
	out = strings.Replace(out, "[ducksoup_audio_test]", "<a href=\"https://ducksoup.psy.gla.ac.uk/test/audio_direct/\" target=\"_blank\">this link</a>", 1)
	out = strings.Replace(out, "[technical_code]", "<input type=\"text\" id=\"technical_code\" name=\"technical_code\"/>", 1)
	out = strings.Replace(out, "[technical_code_alert]", "<div id=\"technical-code-alert-container\" class=\"alert alert-danger\" role=\"alert\">", 1)
	out = strings.Replace(out, "[/technical_code_alert]", "</div>", 1)

	re = regexp.MustCompile(`<a href="(https?://.*?)"[^>]*>(.*?)</a>`)
	out = re.ReplaceAllString(out, `<a href="$1" target="_blank">$2</a>`)

	fmt.Println("Final HTML output:", out) // Debug output

	return template.HTML(out)
}

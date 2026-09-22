package digest

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strings"
	"unicode"

	"github.com/samnodier/jobradar/internal/database"
)

//go:embed email.html
var emailHTML string
var emailTemplate = template.Must(template.New("digest").Funcs(template.FuncMap{
	"initials": func(name string) string {
		words := strings.Fields(name)
		var letters []rune
		for _, word := range words {
			for _, r := range word {
				letters = append(letters, unicode.ToUpper(r))
				break
			}
			if len(letters) == 2 {
				break
			}
		}
		if len(letters) == 0 {
			return "J"
		}
		return string(letters)
	},
	"mode": func(remote *bool) string {
		if remote == nil {
			return ""
		}
		if *remote {
			return "Remote"
		}
		return "On-site"
	},
}).Parse(emailHTML))

type RenderOptions struct{ BaseURL, Schedule string }

func Render(jobs []database.GetDigestCandidatesRow, options ...RenderOptions) (string, string, error) {
	var opts RenderOptions
	if len(options) > 0 {
		opts = options[0]
	}
	data := struct {
		Jobs                           []database.GetDigestCandidatesRow
		Count                          int
		JobsURL, SettingsURL, Schedule string
	}{Jobs: jobs, Count: len(jobs), Schedule: opts.Schedule}
	if opts.BaseURL != "" {
		data.JobsURL = strings.TrimRight(opts.BaseURL, "/") + "/jobs?filter=matched&order=score_desc"
		data.SettingsURL = strings.TrimRight(opts.BaseURL, "/") + "/profile"
	}
	var out bytes.Buffer
	err := emailTemplate.Execute(&out, data)
	return fmt.Sprintf("Jobradar: your top %d job matches", len(jobs)), out.String(), err
}

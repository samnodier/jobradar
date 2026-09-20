package digest

import (
	"bytes"
	"fmt"
	"github.com/samnodier/jobradar/internal/database"
	"html/template"
)

var emailTemplate = template.Must(template.New("digest").Parse(`<!doctype html><html><body style="font-family:system-ui,sans-serif;max-width:680px;margin:auto;padding:24px;color:#172033">
<h1>Your Jobradar matches</h1><p>New roles matching your profile. Scores are a guide, not a guarantee of eligibility.</p>
{{range .}}<article style="padding:16px 0;border-bottom:1px solid #ddd"><h2><a href="{{.SourceUrl}}">{{.Title}}</a></h2><p>{{.CompanyName}} · {{if .JobLocation}}{{.JobLocation}}{{else}}Location unspecified{{end}} · Match {{.MatchScore}}%</p>
{{if eq .JobSource "adzuna"}}<p><a href="https://www.adzuna.com">Jobs</a> by <a href="https://www.adzuna.com"><img src="https://zunastatic-abf.kxcdn.com/images/global/landing/press/logo_normal.png" alt="Adzuna" width="116" height="23"></a></p>{{else}}<p>Source: {{.JobSource}}</p>{{end}}</article>{{end}}
<p>Personal digest from your Jobradar instance. Disable job notifications in Profile → Preferences to stop delivery.</p></body></html>`))

func Render(jobs []database.GetDigestCandidatesRow) (string, string, error) {
	var out bytes.Buffer
	err := emailTemplate.Execute(&out, jobs)
	return fmt.Sprintf("Jobradar: %d new job matches", len(jobs)), out.String(), err
}

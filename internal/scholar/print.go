package scholar

import (
	"bytes"
	"html/template"

	"learning.local/workspace/internal/academic"
)

var booklet = template.Must(template.New("booklet").Parse(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Paper.Reference}} · {{.Paper.Spec.Title}}</title><style>body{max-width:800px;margin:2rem auto;padding:0 1rem;font:16px/1.6 system-ui;color:#182822}h1{line-height:1.2}article{border-top:1px solid #aab6af;margin:1rem 0;padding-top:1rem;break-inside:avoid}img{max-width:100%;max-height:300px}p{white-space:pre-wrap}small{color:#43564c}@page{margin:16mm}@media print{body{margin:0;padding:0;font-size:11pt}a{color:inherit}}</style><h1>{{.Paper.Spec.Title}}</h1><p>{{.Paper.Reference}} · {{.Paper.Spec.DurationMinutes}} minutes · {{.Paper.Maximum}} maximum marks</p><p>{{.Paper.Spec.Instructions}}</p>{{if .Answers}}<h2>Staff marking key · private</h2>{{else}}<p>Use the matching personalized response sheet. Fill the circles completely; leave optional unattempted items blank.</p>{{end}}{{range .Paper.Spec.Sections}}<h2>{{.Title}}</h2><p>Contribution: {{.Policy}} · required count {{.Count}}{{if .Cap}} · section cap {{.Cap}}{{end}}</p>{{range .Questions}}<article><small>{{.Reference}} / pinned revision {{.ID}} · {{.Marks}} marks, wrong-response penalty {{.Penalty}}</small><p>{{.Stem}}</p>{{if .AssetID}}<img src="{{$.Base}}{{.AssetID}}" alt="{{.Alt}}">{{end}}<ol type="A">{{range .Options}}<li>{{.Text}}{{if .AssetID}}<img src="{{$.Base}}{{.AssetID}}" alt="{{.Alt}}">{{end}}</li>{{end}}</ol>{{if $.Answers}}<p><strong>Approved answer rule:</strong> {{range .Key}}{{.}} {{end}}{{if .Numeric}}values {{range .Numeric.Values}}{{.}} {{end}} range {{.Numeric.Min}}–{{.Numeric.Max}} absolute tolerance {{.Numeric.Absolute}} relative tolerance {{.Numeric.Relative}}{{end}}{{range .Rubric}}{{.Description}} / {{.Marks}} marks; {{end}}</p><p>{{.Solution}}</p>{{else}}{{if eq .Type "numerical"}}<p>Numerical response: _____________________</p>{{end}}{{if or (eq .Type "short") (eq .Type "extended")}}<p>Written response: ___________________________________________________________</p><br><br><br>{{end}}{{end}}</article>{{end}}{{end}}<footer><small>Original Learning Workspace output · content and marking policy pinned to the published assessment revision.</small></footer></html>`))

func PrintPaper(p PaperRecord, org string, answers bool) ([]byte, error) {
	if !answers {
		p.Spec = academic.PublicPaper(p.Spec)
	}
	var out bytes.Buffer
	err := booklet.Execute(&out, struct {
		Paper   PaperRecord
		Base    string
		Answers bool
	}{p, "/api/v1/institutions/" + org + "/education/assets/", answers})
	return out.Bytes(), err
}

package scholar

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
)

type OMRTemplateWrite struct {
	Name     string              `json:"name"`
	Template json.RawMessage     `json:"template"`
	Config   json.RawMessage     `json:"config"`
	Identity []OMRIdentityRegion `json:"identity"`
	Layout   json.RawMessage     `json:"layout"`
}
type OMRSavedTemplate struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	CourseID      string    `json:"course_id"` // Private asset provenance, not assessment applicability.
	Origin        string    `json:"created_in"`
	QuestionCount int       `json:"question_count"`
	CreatedAt     time.Time `json:"created_at"`
}
type OMRSheetSummary struct {
	ID                string `json:"id"`
	Page              int    `json:"page"`
	Registration      string `json:"registration_number"`
	Name              string `json:"student_name"`
	State             string `json:"state"`
	Score             string `json:"score"`
	Maximum           string `json:"maximum"`
	Correct           []int  `json:"correct_questions"`
	Wrong             []int  `json:"wrong_questions"`
	Uncertain         []int  `json:"uncertain_questions"`
	IdentityUncertain bool   `json:"identity_uncertain"`
	Version           int    `json:"version"`
}
type OMRCheck struct {
	ID         string            `json:"id"`
	PaperID    string            `json:"paper_id"`
	TemplateID string            `json:"template_id"`
	JobID      string            `json:"job_id"`
	State      string            `json:"state"`
	Progress   int               `json:"progress"`
	Error      string            `json:"error"`
	Filename   string            `json:"filename"`
	CreatedAt  time.Time         `json:"created_at"`
	Sheets     []OMRSheetSummary `json:"sheets"`
}
type OMRCheckedSheet struct {
	Summary OMRSheetSummary            `json:"summary"`
	CheckID string                     `json:"check_id"`
	PaperID string                     `json:"paper_id"`
	AssetID string                     `json:"asset_id"`
	Reading OMRCheckerReading          `json:"reading"`
	Answers map[string]academic.Answer `json:"answers"`
	Result  academic.Result            `json:"result"`
}
type OMRReview struct {
	Version      int                        `json:"version"`
	Registration string                     `json:"registration_number"`
	Name         string                     `json:"student_name"`
	Confirmed    bool                       `json:"confirmed"`
	Reason       string                     `json:"reason"`
	Exclude      bool                       `json:"exclude"`
	Answers      map[string]academic.Answer `json:"answers"`
}
type nativeBlock struct {
	Labels     []string  `json:"fieldLabels"`
	Values     []string  `json:"bubbleValues"`
	Origin     []float64 `json:"origin"`
	Dimensions []float64 `json:"bubbleDimensions"`
	Gap        float64   `json:"bubblesGap"`
	Direction  string    `json:"direction"`
}
type nativeTemplate struct {
	Dimensions []int                  `json:"pageDimensions"`
	Bubbles    []float64              `json:"bubbleDimensions"`
	Blocks     map[string]nativeBlock `json:"fieldBlocks"`
}

var nativeQuestion = regexp.MustCompile(`^q([1-9][0-9]{0,2})$`)

func validateOMRTemplate(in OMRTemplateWrite, reference []byte) (map[int]nativeBlock, error) {
	bad := func() (map[int]nativeBlock, error) {
		return nil, Problem(400, "invalid_omr_template", "Save a complete visual template with a name, matching reference image and consecutive question frames.")
	}
	if !validString(strings.TrimSpace(in.Name), 80) || len(in.Template) > 256<<10 || len(in.Template) == 0 || len(in.Config) > 64<<10 || !json.Valid(in.Config) || len(in.Layout) > 64<<10 || !json.Valid(in.Layout) {
		return bad()
	}
	var config, layout map[string]json.RawMessage
	if strings.ContainsFunc(in.Name, unicode.IsControl) || json.Unmarshal(in.Config, &config) != nil || config == nil || json.Unmarshal(in.Layout, &layout) != nil || layout == nil {
		return bad()
	}
	var definition nativeTemplate
	if json.Unmarshal(in.Template, &definition) != nil || len(definition.Dimensions) != 2 || len(definition.Blocks) < 1 || len(definition.Blocks) > 200 {
		return bad()
	}
	if len(reference) > 8<<20 || len(reference) == 0 {
		return bad()
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(reference))
	if err != nil || (format != "png" && format != "jpeg") || cfg.Width < 100 || cfg.Height < 100 || cfg.Width > 5000 || cfg.Height > 5000 || int64(cfg.Width)*int64(cfg.Height) > 8_000_000 || definition.Dimensions[0] != cfg.Width || definition.Dimensions[1] != cfg.Height {
		return bad()
	}
	if err = validateOMRIdentity(in.Identity); err != nil {
		return nil, err
	}
	fields := map[int]nativeBlock{}
	for _, block := range definition.Blocks {
		if len(block.Labels) != 1 || len(block.Values) < 2 || len(block.Values) > 8 || len(block.Origin) != 2 || block.Direction != "horizontal" {
			return bad()
		}
		match := nativeQuestion.FindStringSubmatch(block.Labels[0])
		if match == nil {
			return bad()
		}
		number, _ := strconv.Atoi(match[1])
		if number > 200 {
			return bad()
		}
		if _, exists := fields[number]; exists {
			return bad()
		}
		digits := block.Values[0] == "1"
		for i, value := range block.Values {
			want := string(rune('A' + i))
			if digits {
				want = strconv.Itoa(i + 1)
			}
			if value != want {
				return bad()
			}
		}
		dims := block.Dimensions
		if len(dims) == 0 {
			dims = definition.Bubbles
		}
		if len(dims) != 2 {
			return bad()
		}
		for _, v := range append(append([]float64{}, block.Origin...), dims...) {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
				return bad()
			}
		}
		if math.IsNaN(block.Gap) || math.IsInf(block.Gap, 0) || block.Gap < 4 || dims[0] < 4 || dims[1] < 4 || dims[0] > 512 || dims[1] > 512 || block.Origin[0]+float64(len(block.Values)-1)*block.Gap+dims[0] > float64(cfg.Width)+1 || block.Origin[1]+dims[1] > float64(cfg.Height)+1 {
			return bad()
		}
		fields[number] = block
	}
	for n := 1; n <= len(fields); n++ {
		if _, ok := fields[n]; !ok {
			return bad()
		}
	}
	return fields, nil
}
func validateOMRIdentity(regions []OMRIdentityRegion) error {
	if len(regions) > 2 {
		return Problem(400, "invalid_identity_regions", "Use at most one registration field and one name field.")
	}
	seen := map[string]bool{}
	for _, region := range regions {
		r := region.Rect
		if (region.Kind != "student_name" && region.Kind != "roll_number") || seen[region.Kind] {
			return Problem(400, "invalid_identity_regions", "Use unique registration and name fields.")
		}
		for _, v := range []float64{r.X, r.Y, r.Width, r.Height} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return Problem(400, "invalid_identity_regions", "Fields must stay inside the sheet.")
			}
		}
		if r.X < 0 || r.Y < 0 || r.Width <= 0 || r.Height <= 0 || r.X+r.Width > 1+1e-8 || r.Y+r.Height > 1+1e-8 {
			return Problem(400, "invalid_identity_regions", "Fields must stay inside the sheet.")
		}
		seen[region.Kind] = true
	}
	return nil
}
func omrQuestions(p academic.Paper) []academic.Question {
	out := []academic.Question{}
	for _, s := range p.Sections {
		out = append(out, s.Questions...)
	}
	return out
}
func matchOMRPaper(p PaperRecord, definition OMRTemplateWrite) error {
	if p.State != "published" {
		return Problem(409, "paper_unpublished", "Publish the reviewed assessment before checking sheets.")
	}
	var t nativeTemplate
	if err := json.Unmarshal(definition.Template, &t); err != nil {
		return err
	}
	questions := omrQuestions(p.Spec)
	if len(t.Blocks) < len(questions) {
		return Problem(400, "omr_template_mismatch", "This template has too few answer rows for this assessment.")
	}
	blocks := map[string]nativeBlock{}
	for _, b := range t.Blocks {
		if len(b.Labels) == 1 {
			blocks[b.Labels[0]] = b
		}
	}
	for i, q := range questions {
		if q.Type != "single" && q.Type != "multiple" {
			return Problem(400, "omr_question_type", "This bubble template needs an assessment containing selection questions.")
		}
		if len(blocks[fmt.Sprintf("q%d", i+1)].Values) != len(q.Options) {
			return Problem(400, "omr_template_mismatch", "Each template question must have the same option count as the assessment.")
		}
	}
	return nil
}
func (s *Service) SaveOMRTemplate(ctx context.Context, tx pgx.Tx, actor, org, course string, in OMRTemplateWrite, reference []byte) (OMRSavedTemplate, error) {
	var campus, origin string
	err := tx.QueryRow(ctx, "SELECT c.campus_id,concat_ws(' · ',c.name,c.academic_year,cp.name) FROM course c JOIN campus cp ON cp.institution_id=c.institution_id AND cp.id=c.campus_id WHERE c.institution_id=$1 AND c.id=$2 AND c.state='active'", org, course).Scan(&campus, &origin)
	if errors.Is(err, pgx.ErrNoRows) {
		return OMRSavedTemplate{}, Problem(404, "course_unavailable", "Choose an available teaching batch for template storage.")
	}
	if err != nil {
		return OMRSavedTemplate{}, err
	}
	if err = Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return OMRSavedTemplate{}, err
	}
	fields, err := validateOMRTemplate(in, reference)
	if err != nil {
		return OMRSavedTemplate{}, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", org+course+strings.ToLower(in.Name)); err != nil {
		return OMRSavedTemplate{}, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM omr_template WHERE institution_id=$1 AND course_id=$2 AND lower(name)=lower($3))", org, course, in.Name).Scan(&exists); err != nil {
		return OMRSavedTemplate{}, err
	}
	if exists {
		return OMRSavedTemplate{}, Problem(409, "template_name_exists", "A template with this name already exists. Choose another name or use the saved template.")
	}
	a, err := s.StoreAsset(ctx, tx, actor, org, campus, course, "scan_page", "omr-reference", reference)
	if err != nil {
		return OMRSavedTemplate{}, err
	}
	out := OMRSavedTemplate{ID: access.Random(), Name: in.Name, CourseID: course, Origin: origin, QuestionCount: len(fields)}
	err = tx.QueryRow(ctx, `INSERT INTO omr_template(id,institution_id,campus_id,course_id,name,question_count,definition,reference_asset_id,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING created_at`, out.ID, org, campus, course, in.Name, len(fields), Encode(in), a.ID, actor).Scan(&out.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return out, Problem(409, "template_name_exists", "A template with this name already exists. Choose another name or use the saved template.")
	}
	if err != nil {
		return out, err
	}
	return out, Event(ctx, tx, org, actor, "omr.template_saved", out.ID, "Named reusable native template")
}
func (s *Service) OMRTemplates(ctx context.Context, tx pgx.Tx, actor, org string) ([]OMRSavedTemplate, error) {
	var campus, course string
	err := tx.QueryRow(ctx, `SELECT c.campus_id,c.id FROM course c JOIN membership m ON m.institution_id=c.institution_id AND m.person_id=$2 AND m.state='active' AND (m.campus_id IS NULL OR m.campus_id=c.campus_id) WHERE c.institution_id=$1 AND c.state='active' AND (m.role IN ('owner','coordinator') OR (m.role='instructor' AND EXISTS(SELECT 1 FROM responsibility r WHERE r.institution_id=c.institution_id AND r.course_id=c.id AND r.staff_id=$2 AND r.starts_on<=current_date AND (r.ends_on IS NULL OR r.ends_on>=current_date)))) LIMIT 1`, org, actor).Scan(&campus, &course)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, Problem(403, "academic_forbidden", "Active teaching access is required for the template library.")
	}
	if err != nil {
		return nil, err
	}
	if err = Scope(ctx, tx, actor, org, campus, course, ""); err != nil {
		return nil, err
	}
	return rows[OMRSavedTemplate](ctx, tx, `SELECT t.id,t.name,t.course_id,concat_ws(' · ',c.name,c.academic_year,cp.name),t.question_count,t.created_at FROM omr_template t JOIN course c ON c.institution_id=t.institution_id AND c.id=t.course_id JOIN campus cp ON cp.institution_id=c.institution_id AND cp.id=c.campus_id WHERE t.institution_id=$1 ORDER BY lower(t.name),t.created_at`, org)
}
func loadOMRTemplate(ctx context.Context, tx pgx.Tx, org, id string) (OMRSavedTemplate, OMRTemplateWrite, string, error) {
	var meta OMRSavedTemplate
	var definition OMRTemplateWrite
	var raw []byte
	var reference string
	err := tx.QueryRow(ctx, `SELECT t.id,t.name,t.course_id,concat_ws(' · ',c.name,c.academic_year,cp.name),t.question_count,t.created_at,t.definition,t.reference_asset_id FROM omr_template t JOIN course c ON c.institution_id=t.institution_id AND c.id=t.course_id JOIN campus cp ON cp.institution_id=c.institution_id AND cp.id=c.campus_id WHERE t.institution_id=$1 AND t.id=$2`, org, id).Scan(&meta.ID, &meta.Name, &meta.CourseID, &meta.Origin, &meta.QuestionCount, &meta.CreatedAt, &raw, &reference)
	if errors.Is(err, pgx.ErrNoRows) {
		return meta, definition, reference, Problem(404, "omr_template_unavailable", "Choose an available saved template.")
	}
	if err == nil {
		err = json.Unmarshal(raw, &definition)
	}
	return meta, definition, reference, err
}
func (s *Service) StageOMRCheck(ctx context.Context, tx pgx.Tx, actor, org, paper, template, filename string, data []byte) (OMRCheck, error) {
	p, err := s.Paper(ctx, tx, actor, org, paper)
	if err != nil {
		return OMRCheck{}, err
	}
	_, definition, _, err := loadOMRTemplate(ctx, tx, org, template)
	if err != nil {
		return OMRCheck{}, err
	}
	if err = matchOMRPaper(p, definition); err != nil {
		return OMRCheck{}, err
	}
	digest := access.Digest(string(data))
	lock := org + paper + template + digest
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", lock); err != nil {
		return OMRCheck{}, err
	}
	var existing string
	err = tx.QueryRow(ctx, "SELECT id FROM omr_check WHERE institution_id=$1 AND paper_id=$2 AND template_id=$3 AND sha256=$4", org, paper, template, digest).Scan(&existing)
	if err == nil {
		return s.OMRCheck(ctx, tx, actor, org, existing)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return OMRCheck{}, err
	}
	asset, err := s.StoreAsset(ctx, tx, actor, org, p.CampusID, p.CourseID, "scan", filename, data)
	if err != nil {
		return OMRCheck{}, err
	}
	if asset.Mime != "application/pdf" {
		cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
		if e != nil || len(data) > 8<<20 || int64(cfg.Width)*int64(cfg.Height) > 8_000_000 {
			return OMRCheck{}, Problem(400, "invalid_omr_image", "Use images of at most 8 MiB and 8 megapixels, or a PDF of at most 20 pages.")
		}
	}
	id, job := access.Random(), access.Random()
	input := map[string]string{"mode": "saved_omr", "paper_id": paper, "check_id": id}
	if _, err = tx.Exec(ctx, `INSERT INTO academic_job(id,institution_id,campus_id,course_id,actor_id,kind,resource_id,input) VALUES($1,$2,$3,$4,$5,'recognize_scan',$6,$7)`, job, org, p.CampusID, p.CourseID, actor, id, Encode(input)); err != nil {
		return OMRCheck{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO omr_check(id,institution_id,paper_id,paper_version,template_id,source_asset_id,sha256,job_id,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, org, paper, p.Version, template, asset.ID, digest, job, actor); err != nil {
		return OMRCheck{}, err
	}
	if err = s.enqueue(ctx, tx, job); err != nil {
		return OMRCheck{}, err
	}
	if err = Event(ctx, tx, org, actor, "omr.check_requested", id, "Registration-based grades; no learner account required"); err != nil {
		return OMRCheck{}, err
	}
	return s.OMRCheck(ctx, tx, actor, org, id)
}
func loadOMRChecks(ctx context.Context, tx pgx.Tx, org, paper, id, before string) ([]OMRCheck, error) {
	r, err := tx.Query(ctx, `SELECT c.id,c.paper_id,c.template_id,c.job_id,j.state,j.progress,j.error,a.filename,c.created_at FROM omr_check c JOIN academic_job j ON j.institution_id=c.institution_id AND j.id=c.job_id JOIN academic_asset a ON a.institution_id=c.institution_id AND a.id=c.source_asset_id WHERE c.institution_id=$1 AND ($2='' OR c.paper_id=$2) AND ($3='' OR c.id=$3) AND ($4='' OR (c.created_at,c.id)<(SELECT created_at,id FROM omr_check WHERE institution_id=$1 AND paper_id=$2 AND id=$4)) ORDER BY c.created_at DESC,c.id DESC LIMIT 100`, org, paper, id, before)
	if err != nil {
		return nil, err
	}
	out := []OMRCheck{}
	ids := []string{}
	byID := map[string]int{}
	for r.Next() {
		var v OMRCheck
		v.Sheets = []OMRSheetSummary{}
		if err = r.Scan(&v.ID, &v.PaperID, &v.TemplateID, &v.JobID, &v.State, &v.Progress, &v.Error, &v.Filename, &v.CreatedAt); err != nil {
			r.Close()
			return nil, err
		}
		byID[v.ID] = len(out)
		ids = append(ids, v.ID)
		out = append(out, v)
	}
	r.Close()
	if err = r.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return out, nil
	}
	r, err = tx.Query(ctx, `SELECT sh.id,sh.check_id,sh.page,sh.registration_number,sh.student_name,sh.state,sh.result->>'score',sh.result->>'maximum',ARRAY(SELECT n::int FROM jsonb_array_elements(sh.result->'items') WITH ORDINALITY item(value,n) WHERE value->>'counted'='true' AND value->>'status'='correct'),ARRAY(SELECT n::int FROM jsonb_array_elements(sh.result->'items') WITH ORDINALITY item(value,n) WHERE value->>'counted'='true' AND value->>'status' IN ('wrong','invalid')),sh.uncertain_questions,sh.identity_uncertain,sh.version FROM omr_checked_sheet sh WHERE sh.institution_id=$1 AND sh.check_id=ANY($2::text[]) ORDER BY sh.page`, org, ids)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	for r.Next() {
		var v OMRSheetSummary
		var check string
		var uncertain []byte
		if err = r.Scan(&v.ID, &check, &v.Page, &v.Registration, &v.Name, &v.State, &v.Score, &v.Maximum, &v.Correct, &v.Wrong, &uncertain, &v.IdentityUncertain, &v.Version); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(uncertain, &v.Uncertain); err != nil {
			return nil, err
		}
		index := byID[check]
		out[index].Sheets = append(out[index].Sheets, v)
	}
	return out, r.Err()
}
func (s *Service) OMRCheck(ctx context.Context, tx pgx.Tx, actor, org, id string) (OMRCheck, error) {
	checks, err := loadOMRChecks(ctx, tx, org, "", id, "")
	if err != nil {
		return OMRCheck{}, err
	}
	if len(checks) == 0 {
		return OMRCheck{}, Problem(404, "omr_check_unavailable", "Sheet upload unavailable.")
	}
	out := checks[0]
	_, err = s.Paper(ctx, tx, actor, org, out.PaperID)
	return out, err
}
func (s *Service) OMRChecks(ctx context.Context, tx pgx.Tx, actor, org, paper, before string) ([]OMRCheck, error) {
	if _, err := s.Paper(ctx, tx, actor, org, paper); err != nil {
		return nil, err
	}
	if before != "" {
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM omr_check WHERE institution_id=$1 AND paper_id=$2 AND id=$3)", org, paper, before).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, Problem(400, "omr_page_unavailable", "Refresh the results list before changing pages.")
		}
	}
	return loadOMRChecks(ctx, tx, org, paper, "", before)
}
func loadOMRSheet(ctx context.Context, tx pgx.Tx, org, id string) (OMRCheckedSheet, error) {
	var out OMRCheckedSheet
	var reading, answers, result, uncertain []byte
	err := tx.QueryRow(ctx, `SELECT id,check_id,paper_id,page,asset_id,registration_number,student_name,state,version,reading,answers,result,uncertain_questions,identity_uncertain FROM omr_checked_sheet WHERE institution_id=$1 AND id=$2`, org, id).Scan(&out.Summary.ID, &out.CheckID, &out.PaperID, &out.Summary.Page, &out.AssetID, &out.Summary.Registration, &out.Summary.Name, &out.Summary.State, &out.Summary.Version, &reading, &answers, &result, &uncertain, &out.Summary.IdentityUncertain)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, Problem(404, "omr_sheet_unavailable", "Checked sheet unavailable.")
	}
	if err != nil {
		return out, err
	}
	for _, v := range []struct {
		raw    []byte
		target any
	}{{reading, &out.Reading}, {answers, &out.Answers}, {result, &out.Result}, {uncertain, &out.Summary.Uncertain}} {
		if err = json.Unmarshal(v.raw, v.target); err != nil {
			return out, err
		}
	}
	out.Summary.Score = out.Result.Score
	out.Summary.Maximum = out.Result.Maximum
	out.Summary.Correct = []int{}
	out.Summary.Wrong = []int{}
	for i, item := range out.Result.Items {
		if item.Counted {
			if item.Status == "correct" {
				out.Summary.Correct = append(out.Summary.Correct, i+1)
			} else if item.Status == "wrong" || item.Status == "invalid" {
				out.Summary.Wrong = append(out.Summary.Wrong, i+1)
			}
		}
	}
	return out, nil
}
func (s *Service) OMRSheet(ctx context.Context, tx pgx.Tx, actor, org, id string) (OMRCheckedSheet, error) {
	out, err := loadOMRSheet(ctx, tx, org, id)
	if err != nil {
		return out, err
	}
	_, err = s.Paper(ctx, tx, actor, org, out.PaperID)
	return out, err
}

func nativeAnswers(p academic.Paper, reading OMRCheckerReading) (map[string]academic.Answer, []int) {
	fields := map[string]OMRCheckerField{}
	for _, f := range reading.Fields {
		fields[f.Label] = f
	}
	answers := map[string]academic.Answer{}
	uncertain := []int{}
	for i, q := range omrQuestions(p) {
		field, exists := fields[fmt.Sprintf("q%d", i+1)]
		a := academic.Answer{Selection: []string{}, Visited: true, Order: i + 1}
		bad := !exists
		seen := map[string]bool{}
		for _, value := range field.Selection {
			option := value
			if n, err := strconv.Atoi(value); err == nil && n >= 1 && n <= len(q.Options) {
				option = q.Options[n-1].ID
			}
			if !slices.ContainsFunc(q.Options, func(o academic.Option) bool { return o.ID == option }) || seen[option] {
				bad = true
				continue
			}
			seen[option] = true
			a.Selection = append(a.Selection, option)
		}
		if q.Type == "single" && len(a.Selection) > 1 {
			a.Invalid = true
			bad = true
		}
		if field.Uncertain && !(q.Type == "multiple" && len(a.Selection) > 1) {
			bad = true
		}
		if field.Value != "" && len(field.Selection) == 0 {
			bad = true
			a.Invalid = true
		}
		a.Review = bad
		if bad {
			uncertain = append(uncertain, i+1)
		}
		answers[q.ID] = a
	}
	return answers, uncertain
}
func checkedIdentity(reading OMRCheckerReading) (registration, name string, uncertain bool) {
	for _, v := range reading.Identity {
		if v.Kind == "roll_number" {
			registration = strings.Join(strings.Fields(v.Text), "")
		} else if v.Kind == "student_name" {
			name = strings.TrimSpace(v.Text)
		}
		uncertain = uncertain || v.Uncertain
	}
	if registration == "" && name == "" {
		uncertain = true
	}
	if len(registration) > 100 || len(name) > 160 || strings.ContainsFunc(registration+name, unicode.IsControl) {
		uncertain = true
	}
	return
}
func registrationConflict(ctx context.Context, tx pgx.Tx, org, paper, registration, except string) (bool, error) {
	if registration == "" {
		return false, nil
	}
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", org+paper+strings.ToLower(registration)); err != nil {
		return false, err
	}
	var found bool
	err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM omr_checked_sheet WHERE institution_id=$1 AND paper_id=$2 AND lower(registration_number)=lower($3) AND id<>$4 AND state='graded')", org, paper, registration, except).Scan(&found)
	return found, err
}
func (s *Service) savedOMRJob(ctx context.Context, j Job, in map[string]string) error {
	if s.Objects == nil {
		return Problem(503, "provider_configuration_required", "Configure private storage before reading retained sheets.")
	}
	var p PaperRecord
	var definition OMRTemplateWrite
	var source, reference Asset
	var check OMRCheck
	err := Tx(ctx, s.DB, func(tx pgx.Tx) error {
		var e error
		check, e = s.OMRCheck(ctx, tx, j.Actor, j.Org, in["check_id"])
		if e != nil {
			return e
		}
		p, e = s.Paper(ctx, tx, j.Actor, j.Org, check.PaperID)
		if e != nil {
			return e
		}
		var ref string
		_, definition, ref, e = loadOMRTemplate(ctx, tx, j.Org, check.TemplateID)
		if e != nil {
			return e
		}
		reference, e = loadAsset(ctx, tx, j.Org, ref)
		if e != nil {
			return e
		}
		var assetID string
		if e = tx.QueryRow(ctx, "SELECT source_asset_id FROM omr_check WHERE institution_id=$1 AND id=$2", j.Org, check.ID).Scan(&assetID); e != nil {
			return e
		}
		source, e = loadAsset(ctx, tx, j.Org, assetID)
		return e
	})
	if err != nil {
		return err
	}
	if err = matchOMRPaper(p, definition); err != nil {
		return err
	}
	data, err := s.Objects.GetAcademic(ctx, source.Key)
	if err != nil {
		return err
	}
	ref, err := s.Objects.GetAcademic(ctx, reference.Key)
	if err != nil {
		return err
	}
	pages, err := rasterPagesLimit(ctx, source.Mime, data, 20)
	if err != nil {
		return err
	}
	for index, page := range pages {
		if err = ctx.Err(); err != nil {
			return err
		}
		var exists bool
		if err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if e := s.jobAllowed(ctx, tx, j); e != nil {
				return e
			}
			return tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM omr_checked_sheet WHERE institution_id=$1 AND check_id=$2 AND page=$3)", j.Org, check.ID, index+1).Scan(&exists)
		}); err != nil {
			return err
		}
		if exists {
			continue
		}
		reading, e := runOMRChecker(ctx, page, nil, ref, definition.Template, definition.Config, definition.Identity)
		if e != nil {
			return e
		}
		answers, uncertain := nativeAnswers(p.Spec, reading)
		registration, name, identityUncertain := checkedIdentity(reading)
		result, e := academic.Mark(p.Spec, answers, nil, "paper")
		if e != nil {
			return e
		}
		overlay, e := base64.StdEncoding.DecodeString(reading.Overlay)
		if e != nil {
			return e
		}
		reading.Overlay = ""
		reading.Notice = "Original OMRChecker responses retained for review. Grades use the selected assessment's pinned answer key and marking rules."
		err = Tx(ctx, s.DB, func(tx pgx.Tx) error {
			if e := s.jobAllowed(ctx, tx, j); e != nil {
				return e
			}
			conflict, e := registrationConflict(ctx, tx, j.Org, p.ID, registration, "")
			if e != nil {
				return e
			}
			identityUncertain = identityUncertain || conflict
			state := "graded"
			if identityUncertain || len(uncertain) > 0 {
				state = "needs_review"
			}
			asset, e := s.StoreAsset(ctx, tx, j.Actor, j.Org, p.CampusID, p.CourseID, "scan_page", fmt.Sprintf("checked-page-%d.jpg", index+1), overlay)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, `INSERT INTO omr_checked_sheet(id,institution_id,check_id,paper_id,page,asset_id,registration_number,student_name,reading,answers,result,uncertain_questions,identity_uncertain,state) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT(institution_id,check_id,page) DO NOTHING`, access.Random(), j.Org, check.ID, p.ID, index+1, asset.ID, registration, name, Encode(reading), Encode(answers), Encode(result), Encode(uncertain), identityUncertain, state)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, "UPDATE academic_job SET progress=$2,provider='OMRChecker',model=$3 WHERE id=$1", j.ID, (index+1)*95/len(pages), reading.Revision)
			return e
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) ReviewOMRSheet(ctx context.Context, tx pgx.Tx, actor, org, id string, in OMRReview) (OMRCheckedSheet, error) {
	if _, err := tx.Exec(ctx, "SELECT id FROM omr_checked_sheet WHERE institution_id=$1 AND id=$2 FOR UPDATE", org, id); err != nil {
		return OMRCheckedSheet{}, err
	}
	sheet, err := s.OMRSheet(ctx, tx, actor, org, id)
	if err != nil {
		return sheet, err
	}
	if sheet.Summary.Version != in.Version {
		return sheet, Problem(409, "omr_sheet_conflict", "Another reviewer changed this sheet. Refresh before saving.")
	}
	if !in.Confirmed || !validString(strings.TrimSpace(in.Reason), 2000) {
		return sheet, Problem(400, "omr_review_required", "Confirm the sheet and record a review reason.")
	}
	registration := strings.Join(strings.Fields(in.Registration), "")
	name := strings.TrimSpace(in.Name)
	if len(registration) > 100 || len(name) > 160 || strings.ContainsFunc(registration+name, unicode.IsControl) || (!in.Exclude && registration == "" && name == "") {
		return sheet, Problem(400, "omr_identity_required", "Enter the registration number or student name visible on this sheet.")
	}
	p, err := s.Paper(ctx, tx, actor, org, sheet.PaperID)
	if err != nil {
		return sheet, err
	}
	answers := make(map[string]academic.Answer, len(sheet.Answers))
	for id, answer := range sheet.Answers {
		answers[id] = answer
	}
	for id, a := range in.Answers {
		q, ok := FindQuestion(p.Spec, id)
		if !ok {
			return sheet, Problem(400, "unknown_question", "An answer refers to another assessment.")
		}
		if !a.Invalid {
			if err = academic.ValidateAnswer(q, a); err != nil {
				return sheet, Invalid(err)
			}
		}
		a.Review = false
		answers[id] = a
	}
	for id, a := range answers {
		a.Review = false
		answers[id] = a
	}
	result, err := academic.Mark(p.Spec, answers, nil, "paper")
	if err != nil {
		return sheet, err
	}
	state := "graded"
	if in.Exclude {
		state = "excluded"
	} else {
		conflict, e := registrationConflict(ctx, tx, org, p.ID, registration, id)
		if e != nil {
			return sheet, e
		}
		if conflict {
			return sheet, Problem(409, "duplicate_registration", "This registration number already has a graded sheet for this assessment. Exclude a duplicate instead of merging students.")
		}
	}
	history := Encode(map[string]any{"version": sheet.Summary.Version, "registration_number": sheet.Summary.Registration, "student_name": sheet.Summary.Name, "answers": sheet.Answers, "result": sheet.Result, "actor_id": actor, "reason": in.Reason, "at": time.Now().UTC()})
	_, err = tx.Exec(ctx, `UPDATE omr_checked_sheet SET registration_number=$3,student_name=$4,answers=$5,result=$6,state=$7,uncertain_questions='[]',identity_uncertain=false,version=version+1,reviews=reviews||jsonb_build_array($8::jsonb) WHERE institution_id=$1 AND id=$2`, org, id, registration, name, Encode(answers), Encode(result), state, history)
	if err != nil {
		return sheet, err
	}
	if err = Event(ctx, tx, org, actor, "omr.sheet_reviewed", id, in.Reason); err != nil {
		return sheet, err
	}
	return s.OMRSheet(ctx, tx, actor, org, id)
}

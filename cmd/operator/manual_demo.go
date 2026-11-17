package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/academic"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/scholar"
)

type manualAccount struct {
	Role     string `json:"role"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Campus   string `json:"campus_id"`
}
type manualManifest struct {
	Institution string          `json:"institution_id"`
	Campus      string          `json:"north_campus_id"`
	SouthCampus string          `json:"south_campus_id"`
	Course      string          `json:"course_id"`
	SouthCourse string          `json:"south_course_id"`
	Accounts    []manualAccount `json:"accounts"`
	CreatedAt   time.Time       `json:"created_at"`
}

func manualDemo(ctx context.Context) error {
	if os.Getenv("APP_ENV") != "development" {
		return errors.New("manual-demo requires APP_ENV=development")
	}
	if data, err := os.ReadFile(".local/manual-access.json"); err == nil {
		var existing manualManifest
		if err = json.Unmarshal(data, &existing); err != nil {
			return err
		}
		if len(existing.Accounts) < 7 {
			return errors.New("manual credential manifest is incomplete; inspect .local/manual-access.json before recovery")
		}
		fmt.Println("Manual test accounts already exist. Credentials and walkthrough: .local/manual-access.md (existing passwords retained).")
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	svc, err := scholar.New(db, nil)
	if err != nil {
		return err
	}
	m := manualManifest{Institution: access.Random(), Campus: access.Random(), SouthCampus: access.Random(), Course: access.Random(), SouthCourse: access.Random(), CreatedAt: time.Now().UTC()}
	for _, role := range []string{"owner", "coordinator", "instructor", "learner", "second-learner", "guardian", "reviewer"} {
		campus := m.Campus
		if role == "owner" || role == "guardian" {
			campus = ""
		}
		if role == "second-learner" {
			campus = m.SouthCampus
		}
		m.Accounts = append(m.Accounts, manualAccount{role, access.Random(), "Manual " + role, "manual." + role + "@demonstration.invalid", access.Random()[:20] + "Aa7!", campus})
	}
	owner := m.Accounts[0].ID
	err = scholar.Tx(ctx, db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('manual-education-demo-v1',0))"); err != nil {
			return err
		}
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM person WHERE email='manual.owner@demonstration.invalid')").Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("manual accounts exist without the private manifest; use account recovery rather than silently resetting passwords")
		}
		if _, err := tx.Exec(ctx, "INSERT INTO institution(id,name,timezone,locale) VALUES($1,'Manual Learning Institute','Asia/Kolkata','en')", m.Institution); err != nil {
			return err
		}
		for _, c := range []struct{ ID, Name, Code string }{{m.Campus, "North Campus", "NORTH"}, {m.SouthCampus, "South Campus", "SOUTH"}} {
			if _, err := tx.Exec(ctx, "INSERT INTO campus(id,institution_id,name,code) VALUES($1,$2,$3,$4)", c.ID, m.Institution, c.Name, c.Code); err != nil {
				return err
			}
		}
		for _, a := range m.Accounts {
			hash, err := access.Password(a.Password)
			if err != nil {
				return err
			}
			if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,$2,$3,$4,'active',now())", a.ID, a.Email, a.Name, hash); err != nil {
				return err
			}
			role := a.Role
			if role == "second-learner" {
				role = "learner"
			}
			var campus any
			if a.Campus != "" {
				campus = a.Campus
			}
			if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,$5)", access.Random(), m.Institution, a.ID, campus, role); err != nil {
				return err
			}
		}
		for _, c := range []struct{ ID, Campus, Name string }{{m.Course, m.Campus, "North foundations"}, {m.SouthCourse, m.SouthCampus, "South foundations"}} {
			if _, err := tx.Exec(ctx, "INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,$4,'2026–27')", c.ID, m.Institution, c.Campus, c.Name); err != nil {
				return err
			}
		}
		for _, a := range m.Accounts {
			if a.Role == "learner" || a.Role == "second-learner" {
				course := m.Course
				if a.Role == "second-learner" {
					course = m.SouthCourse
				}
				if _, err := tx.Exec(ctx, "INSERT INTO enrollment(id,institution_id,person_id,campus_id,course_id,state,start_date) VALUES($1,$2,$3,$4,$5,'active',current_date)", access.Random(), m.Institution, a.ID, a.Campus, course); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, "INSERT INTO family_link(id,institution_id,guardian_id,learner_id,relationship,state,approved_by) VALUES($1,$2,$3,$4,'Synthetic test guardian','approved',$5)", access.Random(), m.Institution, m.Accounts[5].ID, a.ID, owner); err != nil {
					return err
				}
			}
		}
		for _, subject := range []string{"Mathematics", "Physics", "Chemistry", "Biology"} {
			if _, err := tx.Exec(ctx, "INSERT INTO responsibility(id,institution_id,staff_id,course_id,subject,starts_on) VALUES($1,$2,$3,$4,$5,current_date)", access.Random(), m.Institution, m.Accounts[2].ID, m.Course, subject); err != nil {
				return err
			}
		}
		selected := []string{}
		var mathTopic string
		for _, input := range []struct{ Subject, Topic, Outcome, File string }{{"Mathematics", "Addition and divisibility", "Add whole numbers, compute exact decimals and identify divisible numbers.", "questions.csv"}, {"Physics", "Motion and balanced forces", "Calculate constant speed and recognize balanced forces.", "physics.csv"}, {"Chemistry", "Atoms and physical change", "Read simple chemical formulas and distinguish physical from chemical changes.", "chemistry.csv"}, {"Biology", "Cells and food chains", "Identify a membrane's role and primary consumers in a simple food chain.", "biology.csv"}} {
			topic, err := svc.CreateTopic(ctx, tx, owner, m.Institution, scholar.Topic{CampusID: m.Campus, CourseID: m.Course, Subject: input.Subject, Name: input.Topic, Outcome: input.Outcome, Language: "en"})
			if err != nil {
				return err
			}
			if input.Subject == "Mathematics" {
				mathTopic = topic.ID
			}
			data, err := os.ReadFile("apps/web/public/demo/" + input.File)
			if err != nil {
				return err
			}
			parsed, err := academic.ParseImport(input.File, data, topic.ID, "Original manual demonstration material", "CC0-1.0")
			if err != nil {
				return err
			}
			for _, row := range parsed {
				if row.Error != "" {
					return fmt.Errorf("manual content %s row %d: %s", input.File, row.Index, row.Error)
				}
				q, _, err := svc.CreateQuestion(ctx, tx, owner, m.Institution, row.Question, "authored")
				if err != nil {
					return err
				}
				if err = svc.ReviewQuestion(ctx, tx, owner, m.Institution, q.ID, "approved", "Checked original demonstration wording, key, solution and license", 1); err != nil {
					return err
				}
				if input.Subject == "Mathematics" {
					selected = append(selected, q.ID)
				}
			}
		}
		written := academic.Question{TopicID: mathTopic, Type: "short", Stem: "Explain why 12 is divisible by 3, using multiplication or equal groups.", Rubric: []academic.Criterion{{ID: "groups", Description: "Identifies four groups of three or 3 × 4", Marks: "1"}, {ID: "reason", Description: "Explains that there is no remainder", Marks: "1"}}, Solution: "12 = 3 × 4: four equal groups of three use all twelve with no remainder.", Marks: "2", Penalty: "0", Difficulty: "standard", Language: "en", Provenance: "Original manual rubric demonstration", License: "CC0-1.0"}
		q, _, err := svc.CreateQuestion(ctx, tx, owner, m.Institution, written, "authored")
		if err != nil {
			return err
		}
		if err = svc.ReviewQuestion(ctx, tx, owner, m.Institution, q.ID, "approved", "Verified rubric totals and example reasoning", 1); err != nil {
			return err
		}
		writtenID := q.ID
		for _, purpose := range []string{"Online practice", "Paper practice", "Written review practice"} {
			ids := selected
			if purpose == "Written review practice" {
				ids = []string{writtenID}
			}
			paper, err := svc.Compose(ctx, tx, owner, m.Institution, scholar.Compose{CampusID: m.Campus, CourseID: m.Course, Title: "Manual " + purpose, Purpose: "practice", Instructions: "Synthetic manual assessment. Blank answers receive no penalty. Review the per-question marking rules.", DurationMinutes: 45, ReleasePolicy: "immediate", Sections: []scholar.ComposeSection{{ID: "core", Title: "Core evidence", Policy: "all", Count: len(ids), QuestionIDs: ids}}})
			if err != nil {
				return err
			}
			if err = svc.Publish(ctx, tx, owner, m.Institution, paper.ID, 1); err != nil {
				return err
			}
			if _, err = svc.Assign(ctx, tx, owner, m.Institution, paper.ID, []string{m.Accounts[3].ID}, time.Now().Add(-time.Hour), time.Now().AddDate(0, 0, 30), 10); err != nil {
				return err
			}
			if purpose == "Paper practice" {
				if _, err = svc.CreateLayout(ctx, tx, owner, m.Institution, paper.ID, 1); err != nil {
					return err
				}
			}
		}
		return scholar.Event(ctx, tx, m.Institution, owner, "manual_fixture.created", m.Institution, "Seven synthetic accounts; four original subject banks; online, paper and written assessments")
	})
	if err != nil {
		return err
	}
	if err = os.MkdirAll(".local", 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(".local/manual-access.json", append(data, '\n'), 0600); err != nil {
		return err
	}
	var guide strings.Builder
	guide.WriteString("# Manual test accounts\n\nOpen http://localhost:3000/sign-in and select **Manual Learning Institute**. Use separate browser profiles/private windows for separate actors. All identities below are synthetic development accounts.\n\n| Role | Email | Password | Scope |\n|---|---|---|---|\n")
	for _, a := range m.Accounts {
		scope := "North Campus"
		if a.Role == "owner" {
			scope = "Institution-wide"
		}
		if a.Role == "second-learner" {
			scope = "South Campus; boundary checks"
		}
		if a.Role == "guardian" {
			scope = "Family links to both learners; academic family workspace still pending"
		}
		if a.Role == "reviewer" {
			scope = "Account/security only; delegated review workspace still pending"
		}
		fmt.Fprintf(&guide, "| %s | %s | `%s` | %s |\n", a.Role, a.Email, a.Password, scope)
	}
	guide.WriteString("\n## Suggested manual checks\n\n1. Owner: Learning & assessment → Topics & questions. Four subject banks are populated with original approved questions; import more CSV/XLSX/JSON or PDF/image documents, correct previews and approve drafts.\n2. Teacher: the same North course/subject banks are available. South records are outside the teaching assignment. Generate topic drafts with Gemini, inspect Job recovery, review and approve outputs.\n3. Learner: Learning & assessment → Manual Online practice. Answer/flag/clear, interrupt the network, retry saved responses, submit and view the result. A configured contextual tutor is available only after release.\n4. Written review: submit a written answer, then owner/teacher opens Prepare assessment → Manual Written review practice → Review evidence and approves rubric marks.\n5. Coordinator: Paper evidence → Manual Paper practice. Select its registered layout and learner, print the sheet, fill it, create an intake batch, upload a scan, review uncertainty/identity and commit once. Numerical/written areas require manual transcription.\n6. Owner: inspect audit and job recovery. Coordinator is restricted to North Campus; second learner is enrolled only in South. Guardian/reviewer accounts are provided for permission boundaries and account flows; their complete product workspaces are still unimplemented.\n\nPublished assignment windows last 30 days from fixture creation. This command preserves existing passwords and manual work on subsequent runs; it does not silently reset or reopen submitted assessments. Create a fresh paper/assignment in the UI for another attempt. See docs/CORE-EDUCATION.md and IMPLEMENTATION-STATUS.md for the full remaining scope.\n")
	if err = os.WriteFile(".local/manual-access.md", []byte(guide.String()), 0600); err != nil {
		return err
	}
	fmt.Println("Created seven manual accounts and original assessment fixtures. Private credentials and walkthrough: .local/manual-access.md")
	return nil
}

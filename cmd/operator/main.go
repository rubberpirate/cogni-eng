package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"learning.local/workspace/internal/access"
)

func main() {
	if err := run(); err != nil {
		slog.Error("operator failed", "error", err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) < 2 {
		return errors.New("usage: operator migrate|bootstrap|demo")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if os.Args[1] == "migrate" {
		db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				slog.Warn("migration database close failed", "error", closeErr)
			}
		}()
		if err = goose.SetDialect("postgres"); err != nil {
			return err
		}
		return goose.UpContext(ctx, db, "migrations")
	}
	if os.Args[1] == "browser-fixture" {
		return browserFixture(ctx)
	}
	if os.Args[1] != "bootstrap" && os.Args[1] != "demo" {
		return errors.New("unknown operator command")
	}
	demo := os.Args[1] == "demo"
	if demo && os.Getenv("APP_ENV") != "development" {
		return errors.New("demo requires APP_ENV=development")
	}
	email, name, password := os.Getenv("BOOTSTRAP_EMAIL"), os.Getenv("BOOTSTRAP_NAME"), os.Getenv("BOOTSTRAP_PASSWORD")
	if demo {
		email = "owner@demonstration.invalid"
		name = "Demo Coordinator"
		password = access.Random()[:24]
	}
	if email == "" || name == "" {
		return errors.New("set BOOTSTRAP_EMAIL, BOOTSTRAP_NAME and BOOTSTRAP_PASSWORD")
	}
	institutionName, campusName, zone := os.Getenv("BOOTSTRAP_INSTITUTION"), os.Getenv("BOOTSTRAP_CAMPUS"), os.Getenv("BOOTSTRAP_TIMEZONE")
	if demo {
		institutionName = "Demonstration Institute"
		campusName = "River Campus"
		zone = "Asia/Kolkata"
	}
	if institutionName == "" || campusName == "" {
		return errors.New("set BOOTSTRAP_INSTITUTION, BOOTSTRAP_CAMPUS and BOOTSTRAP_TIMEZONE")
	}
	if _, err := time.LoadLocation(zone); err != nil {
		return errors.New("invalid bootstrap timezone")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return errors.New("invalid bootstrap email")
	}
	hash, err := access.Password(password)
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Successful commit closes the transaction; other rollback failures need visibility.
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Warn("transaction rollback failed", "error", rollbackErr)
		}
	}()
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(837201)"); err != nil {
		return err
	}
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM person)").Scan(&exists); err != nil {
		return err
	}
	if exists {
		if demo {
			fmt.Println("Demo already initialized; existing credentials retained.")
			return nil
		}
		return errors.New("bootstrap is one-time; identities already exist")
	}
	id, org, campus := access.Random(), access.Random(), access.Random()
	if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,lower($2),$3,$4,'active',now())", id, email, name, hash); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO institution(id,name,timezone,locale) VALUES($1,$2,$3,'en')", org, institutionName, zone); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO campus(id,institution_id,name,code) VALUES($1,$2,$3,'MAIN')", campus, org, campusName); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,role) VALUES($1,$2,$3,'owner')", access.Random(), org, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO audit_entry(institution_id,actor_id,action,resource_id,reason,request_id) VALUES($1,$2,'institution.bootstrapped',$1,'Explicit operator command','bootstrap')", org, id); err != nil {
		return err
	}
	if demo {
		if err = fixtures(ctx, tx, org, campus); err != nil {
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if demo {
		if err = os.MkdirAll(".local", 0700); err != nil {
			return err
		}
		if err = os.WriteFile(".local/demo-access.txt", []byte("Synthetic development account\nEmail: "+email+"\nPassword: "+password+"\n"), 0600); err != nil {
			return err
		}
		fmt.Println("Demo ready. Local credentials: .local/demo-access.txt")
	}
	return nil
}
func fixtures(ctx context.Context, tx pgx.Tx, org, campus string) error {
	course := access.Random()
	if _, err := tx.Exec(ctx, "INSERT INTO course(id,institution_id,campus_id,name,academic_year) VALUES($1,$2,$3,'Foundations of Science','2026–27')", course, org, campus); err != nil {
		return err
	}
	for _, x := range []struct{ Name, Email, Role string }{{"Asha Example", "asha@demonstration.invalid", "learner"}, {"Neel Example", "neel@demonstration.invalid", "learner"}, {"Mira Example", "mira@demonstration.invalid", "guardian"}, {"Dev Example", "dev@demonstration.invalid", "instructor"}} {
		id := access.Random()
		hash, err := access.Password(access.Random())
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state) VALUES($1,$2,$3,$4,'invited')", id, x.Email, x.Name, hash); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,campus_id,role) VALUES($1,$2,$3,$4,$5)", access.Random(), org, id, campus, x.Role); err != nil {
			return err
		}
	}
	return nil
}

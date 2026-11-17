package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/access"
)

// Browser fixtures use distinct accounts so repeated suites do not weaken sign-in limits.
func browserFixture(ctx context.Context) error {
	if os.Getenv("APP_ENV") != "development" {
		return errors.New("browser fixtures require development mode")
	}
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Successful commit closes the transaction; other rollback failures need visibility.
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Warn("transaction rollback failed", "error", rollbackErr)
		}
	}()
	var org string
	if err = tx.QueryRow(ctx, `SELECT m.institution_id FROM membership m JOIN person p ON p.id=m.person_id WHERE p.email='owner@demonstration.invalid' AND m.role='owner' AND m.state='active' ORDER BY m.created_at LIMIT 1`).Scan(&org); err != nil {
		return errors.New("initialize the synthetic demo before browser tests")
	}
	id, password := access.Random(), access.Random()[:28]
	email := "browser-" + id[:12] + "@demonstration.invalid"
	hash, err := access.Password(password)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO person(id,email,name,password_hash,state,verified_at) VALUES($1,$2,'Browser Fixture',$3,'active',now())", id, email, hash); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO membership(id,institution_id,person_id,role) VALUES($1,$2,$3,'owner')", access.Random(), org, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO audit_entry(institution_id,actor_id,action,resource_id,reason,request_id) VALUES($1,$2,'fixture.created',$2,'Synthetic browser fixture','browser-fixture')", org, id); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	if err = os.MkdirAll(".local", 0700); err != nil {
		return err
	}
	return os.WriteFile(".local/browser-access.txt", []byte("Email: "+email+"\nPassword: "+password+"\n"), 0600)
}

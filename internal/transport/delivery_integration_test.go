//go:build integration

package transport

import (
	"context"
	"os"
	"testing"

	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/delivery"
)

func TestSMTPAcceptanceAndExpiredLeaseRecovery(t *testing.T) {
	address := os.Getenv("TEST_SMTP_ADDR")
	if address == "" {
		t.Skip("TEST_SMTP_ADDR required for local SMTP acceptance")
	}
	h := setup(t)
	ctx := context.Background()
	h.request("POST", h.path("/invitations"), map[string]any{"name": "Mail Fixture", "email": "mail-" + access.Random()[:8] + "@demonstration.invalid", "role": "instructor", "campus_id": h.campus}, 200)
	h.exec("UPDATE delivery_job SET state='running',attempts=1,lease_token='abandoned',lease_until=now()-interval '1 minute'")
	worker := delivery.Worker{DB: h.db, Sender: delivery.SMTP{Address: address, From: "test@workspace.invalid", Local: true}}
	if err := worker.Once(ctx); err != nil {
		t.Fatal(err)
	}
	if err := worker.Once(ctx); err != nil {
		t.Fatal(err)
	}
	var state string
	var attempts int
	var scrubbed bool
	if err := h.db.QueryRow(ctx, "SELECT state,attempts,payload='{}'::jsonb FROM delivery_job").Scan(&state, &attempts, &scrubbed); err != nil {
		t.Fatal(err)
	}
	if state != "accepted" || attempts != 2 || !scrubbed {
		t.Fatal("SMTP acceptance or lease recovery failed", state, attempts, scrubbed)
	}
	var tokens int
	if err := h.db.QueryRow(ctx, "SELECT count(*) FROM security_token").Scan(&tokens); err != nil || tokens != 1 {
		t.Fatal("duplicate security outcome")
	}
}

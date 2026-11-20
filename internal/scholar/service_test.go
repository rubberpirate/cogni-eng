package scholar

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
)

func TestServiceStartsWithLocalAssessmentProvider(t *testing.T) {
	t.Setenv("ACADEMIC_MODEL_PROVIDER", "local")
	t.Setenv("OCR_PROVIDER", "tesseract")
	t.Setenv("OMR_MODEL_PROVIDER", "")
	db, err := pgxpool.New(context.Background(), "postgres://test:test@127.0.0.1:1/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service, err := New(db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := service.Models.(LocalModel); !ok {
		t.Fatal("local provider not selected")
	}
}

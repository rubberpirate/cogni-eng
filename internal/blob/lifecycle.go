package blob

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Store) CleanupOnce(ctx context.Context, db *pgxpool.Pool) error {
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
	var id, key string
	err = tx.QueryRow(ctx, `SELECT a.id,a.object_key FROM avatar_asset a WHERE a.state IN ('staged','retired') AND a.purge_after<now() AND NOT EXISTS(SELECT 1 FROM person p WHERE p.avatar_asset_id=a.id) ORDER BY a.purge_after FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = s.Delete(ctx, key); err != nil {
		if _, saveErr := tx.Exec(ctx, "UPDATE avatar_asset SET last_error='Object deletion failed',purge_after=now()+interval '5 minutes' WHERE id=$1", id); saveErr != nil {
			return saveErr
		}
		return tx.Commit(ctx)
	}
	// Object deletion is idempotent; a crash before commit is safe to retry.
	if _, err = tx.Exec(ctx, "UPDATE avatar_asset SET state='purged',purged_at=now(),last_error='' WHERE id=$1", id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

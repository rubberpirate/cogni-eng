package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/blob"
)

func (s *Server) putAvatar(w http.ResponseWriter, r *http.Request) error {
	var actor Actor
	id := access.Random()
	key := "profile/" + id + ".png"
	if err := s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		actor, err = s.actor(ctx, tx, r, true)
		return err
	}); err != nil {
		return err
	}
	if s.Objects == nil {
		return fail(503, "storage_not_configured", "Profile image storage is not configured.")
	}
	if err := s.limit(r.Context(), "avatar:"+actor.ID, 20); err != nil {
		return err
	}
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, blob.MaxAvatarBytes))
	if err != nil {
		return fail(413, "image_too_large", "Use an image under 4 MiB.")
	}
	data, err := blob.NormalizeAvatar(raw)
	if err != nil {
		return fail(400, "invalid_image", err.Error())
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.actor(ctx, tx, r, true)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, "INSERT INTO avatar_asset(id,person_id,object_key,state,bytes,sha256,purge_after) VALUES($1,$2,$3,'staged',$4,$5,now()+interval '1 hour')", id, a.ID, key, len(data), access.Digest(string(data)))
		return err
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err = s.Objects.Put(ctx, key, data); err != nil {
		if _, markErr := s.DB.Exec(r.Context(), "UPDATE avatar_asset SET last_error='Upload did not complete',purge_after=now() WHERE id=$1 AND state='staged'", id); markErr != nil {
			return markErr
		}
		return fail(503, "storage_unavailable", "The image was not saved. Your previous image is unchanged; retry later.")
	}
	err = s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.actor(ctx, tx, r, true)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "SELECT id FROM person WHERE id=$1 FOR UPDATE", a.ID); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE avatar_asset SET state='retired',purge_after=now() WHERE id=(SELECT avatar_asset_id FROM person WHERE id=$1)", a.ID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, "UPDATE avatar_asset SET state='active',purge_after=NULL WHERE id=$1 AND person_id=$2 AND state='staged'", id, a.ID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return fail(409, "upload_expired", "This upload expired. Save the retained crop again.")
		}
		if _, err = tx.Exec(ctx, "UPDATE person SET avatar_asset_id=$2 WHERE id=$1", a.ID, id); err != nil {
			return err
		}
		return audit(ctx, tx, r, a, "", "profile.image_updated", id, "")
	})
	if err != nil {
		return err
	}
	write(w, 200, map[string]string{"state": "saved"})
	return nil
}
func (s *Server) getAvatar(w http.ResponseWriter, r *http.Request) error {
	var key string
	err := s.transaction(r, func(ctx context.Context, tx pgx.Tx) error {
		a, err := s.actor(ctx, tx, r, false)
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, "SELECT f.object_key FROM person p JOIN avatar_asset f ON f.id=p.avatar_asset_id WHERE p.id=$1 AND f.person_id=$1 AND f.state='active'", a.ID).Scan(&key)
		if errors.Is(err, pgx.ErrNoRows) {
			return fail(404, "no_profile_image", "No profile image saved.")
		}
		return err
	})
	if err != nil {
		return err
	}
	if s.Objects == nil {
		return fail(503, "storage_not_configured", "Profile image storage is not configured.")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	data, err := s.Objects.Get(ctx, key)
	if err != nil {
		return fail(503, "storage_unavailable", "Profile image unavailable. Retry later.")
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(200)
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}
func (s *Server) deleteAvatar(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		if _, err := tx.Exec(ctx, "SELECT id FROM person WHERE id=$1 FOR UPDATE", a.ID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE avatar_asset SET state='retired',purge_after=now() WHERE id=(SELECT avatar_asset_id FROM person WHERE id=$1)", a.ID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, "UPDATE person SET avatar_asset_id=NULL WHERE id=$1", a.ID); err != nil {
			return nil, err
		}
		err := audit(ctx, tx, r, a, "", "profile.image_removed", a.ID, "")
		return map[string]string{"state": "removed"}, err
	})
}

func (s *Server) avatarHistory(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		type Asset struct {
			ID        string    `json:"id"`
			State     string    `json:"state"`
			Bytes     int       `json:"bytes"`
			CreatedAt time.Time `json:"created_at"`
			Error     string    `json:"error"`
		}
		return collect[Asset](ctx, tx, "SELECT id,state,bytes,created_at,last_error FROM avatar_asset WHERE person_id=$1 ORDER BY created_at DESC LIMIT 30", a.ID)
	})
}
func (s *Server) retryAvatarCleanup(w http.ResponseWriter, r *http.Request) error {
	return s.authenticated(w, r, func(ctx context.Context, tx pgx.Tx, a Actor) (any, error) {
		tag, err := tx.Exec(ctx, "UPDATE avatar_asset SET purge_after=now(),last_error='' WHERE id=$1 AND person_id=$2 AND (state='retired' OR (state='staged' AND (last_error<>'' OR created_at<now()-interval '1 hour'))) AND NOT EXISTS(SELECT 1 FROM person WHERE avatar_asset_id=$1)", r.PathValue("id"), a.ID)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 0 {
			return nil, fail(409, "invalid_state", "Only an unused image can be scheduled for cleanup.")
		}
		err = audit(ctx, tx, r, a, "", "profile.image_cleanup_retried", r.PathValue("id"), "User request")
		return map[string]string{"state": "queued"}, err
	})
}

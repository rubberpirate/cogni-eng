//go:build integration

package transport

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"learning.local/workspace/internal/access"
	"learning.local/workspace/internal/blob"
)

func TestPrivateAvatarStorageAndLifecycle(t *testing.T) {
	if os.Getenv("OBJECT_ENDPOINT") == "" {
		t.Skip("OBJECT_ENDPOINT required for actual S3 integration")
	}
	h := setup(t)
	cookie := h.cookie
	h.cookie = nil
	h.request("PUT", "/api/v1/profile/avatar", nil, 401)
	h.cookie = cookie
	h.request("PUT", "/api/v1/profile/avatar", nil, 503)
	store, err := blob.FromEnv()
	if err != nil || store == nil {
		t.Fatal("S3 configuration required", err)
	}
	server := New(h.db, Config{Env: "test", Origin: "http://localhost:3000", MFAKey: make([]byte, 32)})
	server.Objects = store
	h.handler = server.Handler()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	img.Set(0, 0, color.RGBA{R: 100, G: 180, A: 255})
	var data bytes.Buffer
	if err = png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	upload := func(content []byte, want int) {
		t.Helper()
		req := httptest.NewRequest("PUT", "/api/v1/profile/avatar", bytes.NewReader(content))
		req.AddCookie(h.cookie)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-CSRF-Token", h.csrf)
		req.Header.Set("Content-Type", "image/png")
		w := httptest.NewRecorder()
		h.handler.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("upload %d: %s", w.Code, w.Body.String())
		}
	}
	upload(data.Bytes(), 200)
	var freshStagedID string
	h.exec("INSERT INTO avatar_asset(id,person_id,object_key,state,bytes,sha256,purge_after) VALUES('fresh-staging',$1,'profile/fresh-staging-not-uploaded.png','staged',64,'synthetic',now()+interval '1 hour')", h.owner)
	freshStagedID = "fresh-staging"
	h.request("POST", "/api/v1/profile/image-history/"+freshStagedID+"/retry", nil, 409)
	var key, id string
	ctx := context.Background()
	if err = h.db.QueryRow(ctx, "SELECT object_key,id FROM avatar_asset WHERE person_id=$1 AND state='active'", h.owner).Scan(&key, &id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Delete(context.Background(), key); err != nil {
			t.Error(err)
		}
	})
	response, err := http.Get(os.Getenv("OBJECT_ENDPOINT") + "/" + store.Bucket + "/" + key)
	if err != nil {
		t.Fatal(err)
	}
	if err = response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 403 {
		t.Fatal("anonymous S3 access was not denied", response.StatusCode)
	}
	req := httptest.NewRequest("GET", "/api/v1/profile/avatar", nil)
	req.AddCookie(h.cookie)
	w := httptest.NewRecorder()
	h.handler.ServeHTTP(w, req)
	if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Type") != "image/png" {
		t.Fatal("private image read failed", w.Code)
	}
	req = httptest.NewRequest("GET", "/api/v1/profile/avatar", nil)
	w = httptest.NewRecorder()
	h.handler.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatal("unauthenticated image accessible")
	}
	upload([]byte("<svg onload='alert(1)'/>"), 400)
	var retained string
	if err = h.db.QueryRow(ctx, "SELECT avatar_asset_id FROM person WHERE id=$1", h.owner).Scan(&retained); err != nil || retained != id {
		t.Fatal("invalid upload destroyed saved image")
	}
	faulty, err := minio.New("127.0.0.1:1", &minio.Options{Creds: credentials.NewStaticV4("synthetic", "synthetic", ""), Secure: false, Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	server.Objects = &blob.Store{Client: faulty, Bucket: store.Bucket}
	upload(data.Bytes(), 503)
	server.Objects = store
	if err = h.db.QueryRow(ctx, "SELECT avatar_asset_id FROM person WHERE id=$1", h.owner).Scan(&retained); err != nil || retained != id {
		t.Fatal("storage failure destroyed previous image")
	}
	ownerCookie, ownerCSRF := h.cookie, h.csrf
	other := access.Random()
	hash, _ := access.Password("a second synthetic image owner!")
	h.exec("INSERT INTO person(id,email,name,password_hash,state) VALUES($1,'image-other@example.invalid','Other Person',$2,'active')", other, hash)
	h.login("image-other@example.invalid", "a second synthetic image owner!", "")
	h.request("GET", "/api/v1/profile/avatar?person_id="+h.owner, nil, 404)
	h.cookie, h.csrf = ownerCookie, ownerCSRF
	h.request("DELETE", "/api/v1/profile/avatar", nil, 200)
	if err = store.CleanupOnce(ctx, h.db); err != nil {
		t.Fatal(err)
	}
	if err = store.CleanupOnce(ctx, h.db); err != nil {
		t.Fatal(err)
	}
	var state string
	if err = h.db.QueryRow(ctx, "SELECT state FROM avatar_asset WHERE id=$1", id).Scan(&state); err != nil || state != "purged" {
		t.Fatal("cleanup did not persist", state, err)
	}
	if _, err = store.Get(ctx, key); err == nil {
		t.Fatal("deleted private image still retrievable")
	}
}

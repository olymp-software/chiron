package library

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func serve(h http.Handler, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestFilesHandler(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, "images", "setup.PNG"), "png-bytes")
	writeFile(t, filepath.Join(src, "notes.txt"), "text")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret.png"), "secret")
	linked := os.Symlink(filepath.Join(outside, "secret.png"), filepath.Join(l.CopyPath(e.Key), "linked.png")) == nil
	h := l.FilesHandler()
	base := FilesPrefix + e.Key + "/"

	rec := serve(h, http.MethodGet, base+"images/setup.PNG")
	if rec.Code != http.StatusOK || rec.Body.String() != "png-bytes" {
		t.Fatalf("GET image = %d %q, want 200 with the file", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}

	rejected := []string{
		base + "notes.txt",
		base + "course.yaml",
		base + "../" + e.Key + "/images/setup.PNG",
		base + "images/../../library.json.png",
		base + "missing.png",
		FilesPrefix + "unknown-key/images/setup.PNG",
		FilesPrefix + e.Key,
		"/other/" + e.Key + "/images/setup.PNG",
	}
	if linked {
		rejected = append(rejected, base+"linked.png")
	}
	for _, target := range rejected {
		if rec := serve(h, http.MethodGet, target); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, rec.Code)
		}
	}
	if rec := serve(h, http.MethodPost, base+"images/setup.PNG"); rec.Code != http.StatusNotFound {
		t.Errorf("POST = %d, want 404", rec.Code)
	}
}

func TestFilesHandlerAbsolutePath(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret.png"), "secret")

	// The handler's path is joined into the course copy, never used as is.
	target := FilesPrefix + e.Key + "/" + filepath.ToSlash(filepath.Join(outside, "secret.png"))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Path = target
	rec := httptest.NewRecorder()
	l.FilesHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET %s = %d, want 404", target, rec.Code)
	}
}

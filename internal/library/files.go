package library

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FilesPrefix is the URL prefix under which course files are served to the
// webview: /course-files/<key>/<path in the course>.
const FilesPrefix = "/course-files/"

// imageTypes are the only files served, by lowercase extension. SVG is shown
// through <img>, where its scripts do not run.
var imageTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
}

// FilesHandler serves images from course copies. Paths that leave the course
// copy, including through symbolic links, are not served.
func (l *Library) FilesHandler() http.Handler {
	return http.HandlerFunc(l.serveFile)
}

func (l *Library) serveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	rest, ok := strings.CutPrefix(r.URL.Path, FilesPrefix)
	if !ok {
		http.NotFound(w, r)
		return
	}
	key, name, ok := strings.Cut(rest, "/")
	contentType := imageTypes[strings.ToLower(path.Ext(name))]
	if !ok || contentType == "" || !l.hasKey(key) {
		http.NotFound(w, r)
		return
	}

	root, err := os.OpenRoot(l.CopyPath(key))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = root.Close() }()
	// os.Root rejects absolute paths, .. and links leading outside the root.
	f, err := root.Open(filepath.FromSlash(name))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (l *Library) hasKey(key string) bool {
	courses, err := l.List()
	if err != nil {
		return false
	}
	for _, e := range courses {
		if e.Key == key {
			return true
		}
	}
	return false
}

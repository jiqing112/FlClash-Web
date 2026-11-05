package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// The frontend bundle is copied into internal/api/dist by the build script
// and embedded, so one binary serves everything.
//
//go:embed all:dist
var distFS embed.FS

// StaticHandler serves the SPA with an index fallback.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		_, statErr := fs.Stat(sub, path)
		if path == "" || statErr != nil || strings.HasSuffix(path, ".html") {
			// The HTML shell references hashed assets; serve it fresh so a
			// new build is picked up on reload. Assets are hashed, so they
			// can cache heuristically.
			w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		}
		if statErr != nil {
			// SPA fallback
			r2 := new(http.Request)
			*r2 = *r
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

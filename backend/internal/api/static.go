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
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(sub, path); err != nil {
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

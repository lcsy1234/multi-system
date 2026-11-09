// Package web serves the independently built React application.
package web

import (
	"net/http"
	"os"
)

// Run from mes/ by default; MES_WEB_DIR may point to a deployed build directory.
// The Go binary no longer embeds frontend sources, so frontend rebuilds need no Go rebuild.
func Handler() http.Handler {
	directory := os.Getenv("MES_WEB_DIR")
	if directory == "" {
		directory = "../mes-web/dist"
	}
	files := http.FileServer(http.Dir(directory))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Cache-Control", "no-cache")
			if _, err := os.Stat(directory + "/index.html"); err != nil {
				http.Error(w, "MES frontend is not built. Run make frontend-build, or use the Vite dev server on port 5173.", http.StatusServiceUnavailable)
				return
			}
		}
		files.ServeHTTP(w, r)
	})
}

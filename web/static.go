package web

import (
	"io/fs"
	"log"
	"net/http"
)

func getStaticFilesHandler(publicFiles fs.FS) http.Handler {
	staticFS := fs.FS(publicFiles)

	htmlContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(htmlContent)))
}

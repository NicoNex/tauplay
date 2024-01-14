package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed assets
var assets embed.FS

func main() {
	root, err := fs.Sub(assets, "assets")
	if err != nil {
		log.Fatal("fs.Sub", err)
	}

	http.Handle("/", http.FileServer(http.FS(root)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

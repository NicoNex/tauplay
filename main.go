package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/NicoNex/katalis"
)

type Page struct {
	Code    string
	Created time.Time
	Updated time.Time
}

var (
	//go:embed assets
	assets embed.FS
	db     katalis.DB[string, Page]
)

func handleGetCode(w http.ResponseWriter, r *http.Request) {

}

func handleSetCode(w http.ResponseWriter, r *http.Request) {

}

func retry(d time.Duration, fn func()) {
	for {
		fn()
		time.Sleep(d)
	}
}

func serve(port string) {
	root, err := fs.Sub(assets, "assets")
	if err != nil {
		log.Fatal("fs.Sub", err)
	}

	http.Handle("/", http.FileServer(http.FS(root)))
	http.HandleFunc("GET /p/{id}", handleGetCode)
	http.HandleFunc("PUT /p/{id}", handleSetCode)

	retry(time.Second*5, func() {
		log.Println(http.ListenAndServe(port, nil))
	})
}

func initDB() {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Fatalln("initDB", "os.UserCacheDir", err)
	}

	db, err = katalis.Open(
		filepath.Join(cache, "tauplay"),
		katalis.StringCodec,
		katalis.GobCodec[Page]{},
	)
	if err != nil {
		log.Fatalln("initDB", "katalis.Open", err)
	}
}

func main() {
	port := *flag.String("p", ":8080", "The port TauPlay will listen to.")
	flag.Parse()

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	initDB()
	go serve(port)

	var signals = make(chan os.Signal, 1)

	defer db.Close()
	signal.Notify(signals, syscall.SIGINT, syscall.SIGABRT)
	<-signals
}

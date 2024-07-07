package main

import (
	"embed"
	"flag"
	"io"
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
	"github.com/google/uuid"
)

type Page struct {
	Code    string
	Created time.Time
	Read    time.Time
}

var (
	//go:embed assets
	assets embed.FS
	db     katalis.DB[string, Page]
)

func handleGetCode(w http.ResponseWriter, r *http.Request) {
	p, err := db.Get(r.PathValue("id"))
	if err != nil {
		log.Println("handleGetCode", "db.Get", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func(p Page) {
		p.Read = time.Now()
		if err := db.Put(r.PathValue("id"), p); err != nil {
			log.Println("handleGetCode", "db.Put", err)
			return
		}
	}(p)

	if _, err := io.WriteString(w, p.Code); err != nil {
		log.Println("handleGetCode", "io.WriteString", err)
		return
	}
}

func handleSetCode(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("handleSetCode", "io.ReadAll", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	now := time.Now()
	id := uuid.New().String()
	err = db.Put(id, Page{
		Code:    string(b),
		Created: now,
		Read:    now,
	})
	if err != nil {
		log.Println("handleSetCode", "db.Put", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.WriteString(w, id); err != nil {
		log.Println("handleSetCode", "io.WriteString", err)
		return
	}
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
	http.HandleFunc("PUT /p", handleSetCode)
	http.HandleFunc("GET /p/{id}", handleGetCode)

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
	defer db.Close()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGABRT)
	<-sigs
}

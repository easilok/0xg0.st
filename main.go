package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
)

var port *uint64
var tmpl *template.Template
var host *string

// Dead simple router that just does the **perform** the job
func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.Header.Get("Content-type"), "multipart/form-data"):
		upload(w, r)
	case uuidMatch.MatchString(r.URL.Path):
		getFile(w, r)
	default:
		home(w, r)
	}
}

// Route handling, logging and application serving
func main() {
	// Random seed creation
	rand.Seed(time.Now().Unix())

	// Home template initalization
	tmpl = template.Must(template.ParseFiles("./templates/index.html"))
	// Flags for the leveled logging

	host = flag.String("h", "0.0.0.0", "Address to serve on")
	port = flag.Uint64("p", 8000, "port")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE: ./0xg0.st -p=8080 -stderrthreshold=[INFO|WARNING|FATAL] -log_dir=[string]\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	glog.Flush()

	glog.Infof(`Starting Server on port "%d"`, port)

	// Routing
	http.HandleFunc("/", router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil))
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/coffemanfp/trace"
)

func main() {

	var port = flag.Int("port", 8080, "The port of the application.")
	flag.Parse()

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	// Root
	http.Handle("/", &templateHandler{
		filename: "chat.html",
	})

	http.Handle("/room", r)

	// get the room going
	go r.run()

	log.Println("Starting web server on:", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

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
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
)

func main() {

	var port = flag.Int("port", 8080, "The port of the application.")

	flag.Parse()

	gomniauth.SetSecurityKey("some long key")
	gomniauth.WithProviders(
		facebook.New(
			"274812483806526",
			"2c080df91ae46b808771f9fb9cd750ab",
			"http://localhost:8080/auth/callback/facebook",
		),
		github.New(
			"8ea52079890166be3b8a",
			"7307d137bf9d857f4bc04ede45e1fa65220afe65",
			"http://localhost:8080/auth/callback/github",
		),
	)

	r := newRoom(UseGravatar)
	r.tracer = trace.New(os.Stdout)

	// Assets
	http.Handle(
		"/assets/",
		http.StripPrefix("/assets",
			http.FileServer(
				http.Dir("assets/"),
			),
		),
	)

	// Root
	http.Handle("/chat", MustAuth(&templateHandler{
		filename: "chat.html",
	}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header()["Location"] = []string{"/chat"}
		w.WriteHeader(http.StatusTemporaryRedirect)
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

	data := map[string]interface{}{
		"Host": r.Host,
	}

	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

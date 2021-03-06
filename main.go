package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
)

type User struct {
	ID    string
	Notes []Note
}

type Note struct {
	UserID   string
	ID       string
	Deleted  bool
	Data     string
	Children []Note
}

func main() {
	var users sync.Map

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	http.HandleFunc("/v1/user", func(w http.ResponseWriter, r *http.Request) {
		var user User

		if r.Method == http.MethodGet {
			found, ok := users.Load(r.URL.Query().Get("userID"))
			if ok {
				user = found.(User)
			}
		}

		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				log.Print(err)
				http.Error(w, "Unable to decode user", http.StatusBadRequest)
				return
			} else {
				users.Store(user.ID, user)
			}
		}

		if err := json.NewEncoder(w).Encode(&user); err != nil {
			log.Print(err)
		}
	})

	http.HandleFunc("/service-worker.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/service-worker.js")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		production := r.URL.Query().Get("production") == "true"

		if strings.Contains(r.URL.Path, "/public") {
			path := fmt.Sprintf(".%s", r.URL.Path)
			log.Printf("Serving static file: %s", path)
			http.ServeFile(w, r, path)
			return
		}

		var t = template.Must(template.New("main.template.html").ParseFiles("main.template.html"))

		var css []byte
		var js []byte

		dir := os.DirFS("./")
		cssGlob, err := fs.Glob(dir, "*.css")
		if err != nil {
			panic(err)
		}

		jsGlob, err := fs.Glob(dir, "*.js")
		if err != nil {
			panic(err)
		}

		for _, file := range cssGlob {
			b, err := fs.ReadFile(dir, file)
			if err != nil {
				panic(err)
			}

			css = append(css, b...)
		}

		if production {
			css, err = m.Bytes("text/css", css)
			if err != nil {
				panic(err)
			}
		}

		for _, file := range jsGlob {
			if strings.Contains(file, "test") && production {
				continue
			}

			b, err := fs.ReadFile(dir, file)
			if err != nil {
				panic(err)
			}

			js = append(js, b...)
		}

		if production {
			js, err = m.Bytes("application/javascript", js)
			if err != nil {
				panic(err)
			}
		}

		type data struct {
			CSS []byte
			JS  []byte
		}

		var writer io.Writer
		writer = w
		if production {
			mw := m.Writer("text/html", w)
			defer mw.Close()
			writer = mw
		}

		if err := t.Execute(writer, &data{
			CSS: css,
			JS:  js,
		}); err != nil {
			log.Print(err)
		}
	})

	log.Print("Listening on port :", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil))
}

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

const urlListing = `<pre>{{range $i, $u :=  .}}<a href="{{$u}}">{{$u}}</a>{{end}}</pre>`

var (
	templates *template.Template
)

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v\t%v\t%v\n", r.RemoteAddr, r.Method, r.URL.String())
		h.ServeHTTP(w, r)
	})
}

func init() {
	var err error
	templates, err = template.New("urls").Parse(urlListing)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	laddr := flag.String("l", ":8888", "listening address")
	flag.Parse()

	var paths []string

	for _, dir := range flag.Args() {
		dir = filepath.Clean(dir) + "/"
		pathname := dir
		http.Handle(pathname, http.StripPrefix(pathname, http.FileServer(http.Dir(dir))))
		paths = append(paths, pathname)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := templates.ExecuteTemplate(w, "urls", paths)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code)+": "+err.Error(), code)
		}
	})
	err := http.ListenAndServe(*laddr, logHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bye!")
}

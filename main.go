package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const urlListing = `<pre>{{range $i, $u :=  .}}<a href="{{$u}}">{{$u}}</a>
{{end}}</pre>`

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

func parseArg(arg string) (route string, fpath string, err error) {
	chunks := strings.Split(arg, ":")
	if len(chunks) != 2 {
		err = fmt.Errorf("malformed argument: %s", arg)
		return
	}

	route = path.Clean("/" + chunks[0])
	fpath = filepath.Clean(chunks[1])

	fi, err := os.Stat(fpath)
	if err != nil {
		return
	}

	if !fi.IsDir() {
		err = fmt.Errorf("argument is not a dir (%s in %s)", fpath, arg)
		return
	}
	if route != "/" {
		route = route + "/"
	}
	return
}

func main() {
	laddr := flag.String("l", ":8888", "listening address")
	flag.Parse()

	var (
		routes   []string
		needRoot = true
	)

	for _, a := range flag.Args() {
		route, fpath, err := parseArg(a)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Serving", fpath, "under", route)
		addHandle(route, fpath)

		routes = append(routes, route)
		if route == "/" {
			needRoot = false
		}
	}

	if needRoot {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err := templates.ExecuteTemplate(w, "urls", routes)
			if err != nil {
				code := http.StatusInternalServerError
				http.Error(w, http.StatusText(code)+": "+err.Error(), code)
			}
		})
	}

	err := http.ListenAndServe(*laddr, logHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Bye!")
}

func addHandle(route, fpath string) {
	defer func() {
		rec := recover()
		if rec != nil {
			log.Fatalf("Error, %v", rec)
		}
	}()

	http.Handle(route, http.StripPrefix(route, http.FileServer(http.Dir(fpath))))
}

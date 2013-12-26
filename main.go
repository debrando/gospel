package main

import (
	"flag"
	"html/template"
	"net/http"
	"strconv"
)

var g_port int
var g_tmpls *template.Template

func init() {
	// parameters
	flag.IntVar(&g_port, "port", 8088, "Listening port")
	flag.Parse()
	// templates
	g_tmpls = template.Must(template.ParseFiles(tmplFName("default")))
}

func main() {
	http.Handle("/sts/", http.FileServer(http.Dir("./res/")))
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(g_port), nil))
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "default")
}

func tmplFName(tmpl string) string {
	n := "./res/templates/" + tmpl + ".html"
	return n
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := g_tmpls.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

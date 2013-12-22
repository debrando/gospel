package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var port int
var tmpls *template.Template

func init() {
	// parameters
	flag.IntVar(&port, "port", 8088, "Listening port")
	flag.Parse()
	// templates
	tmpls = template.Must(template.ParseFiles(tmplFName("default")))
}

func main() {
	http.Handle("/sts/", http.FileServer(http.Dir("./res/")))
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(port), nil))
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
	log.Println("Rendering " + tmpl)
	err := tmpls.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

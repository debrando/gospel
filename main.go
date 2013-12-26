package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type jsonResponse map[string]interface{}

func (r jsonResponse) String() (s string) {
	b, err := json.Marshal(r)
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}

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
	http.HandleFunc("/json/", jsonHandler)
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(g_port), nil))
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "default")
}

// JSONs Handlers
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, jsonResponse{"success": true, "message": "Hello!"})
	return
}

// Create full name of template
func tmplFName(tmpl string) string {
	n := "./res/templates/" + tmpl + ".html"
	return n
}

// Render a template by name
func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := g_tmpls.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

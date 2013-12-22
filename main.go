package main 

import (
	"html/template"
    "net/http"
    "flag"
    "strconv"
)

var home string
var port int
var tmpls *template.Template

func init() {
	// parameters
	flag.StringVar(&home, "home", "/usr/local/gospel", "Installing path")
	flag.IntVar(&port, "port", 8088, "Listening port")
	flag.Parse()	
	// templates
	tmpls = template.Must(template.ParseFiles(tmplFName("default")))
}

func main() {
    http.Handle("/res", http.FileServer(http.Dir(home + "/res/sts")))
    http.HandleFunc("/", defaultHandler)
    panic(http.ListenAndServe(":" + strconv.Itoa(port), nil))
}


// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
 	renderTemplate(w, "default")
}
 
func tmplFName(tmpl string) string {
	return home + "/res/templates/" + tmpl + ".html"
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
    err := tmpls.ExecuteTemplate(w, tmplFName(tmpl), nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

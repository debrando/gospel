package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
)

const TEXTHTML = "text/html; charset=utf-8"
const APPJSON = "application/json; charset=utf-8"

type BaseMsg struct {
	Success bool
	Message string
}

var g_port int
var g_tmpls *template.Template
var g_mgos *mgo.Session

func init() {
	// parameters
	var mgourl string
	var err error
	flag.IntVar(&g_port, "port", 8088, "Listening port")
	flag.StringVar(&mgourl, "mongourl", "127.0.0.1:27017", "Mongodb URL")
	flag.Parse()
	// templates
	g_tmpls = template.Must(template.ParseFiles(tmplFName("default")))
	g_mgos, err = mgo.Dial(mgourl + "/gospel")
	if err != nil {
		panic(err)
	}
}

func main() {
	defer g_mgos.Close()
	http.Handle("/sts/", http.FileServer(http.Dir("./res/")))
	http.HandleFunc("/msg/", msgHandler)
	http.HandleFunc("/", defaultHandler)
	panic(http.ListenAndServe(":"+strconv.Itoa(g_port), nil))
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "default")
}

// Message Handlers
func msgHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", APPJSON)
	c := g_mgos.DB("").C("messages")
	var result []BaseMsg
	err := c.Find(bson.M{"success": true}).Sort("-ts").Limit(100).Iter().All(&result)
	if err != nil {
		panic(err)
	}
	b, err := json.Marshal(result)
	fmt.Fprint(w, string(b))
	if err != nil {
		panic(err)
	}
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

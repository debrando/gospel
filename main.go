package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
)

const TEXTHTML = "text/html; charset=utf-8"
const APPJSON = "application/json; charset=utf-8"
const LOCADDRESS = "http://127.0.0.1:8088"
const HKADDRESS = "http://gospel99.herokuapp.com"

type BaseMsg struct {
	Success bool
	Message string
}

var g_tmpls *template.Template
var g_mgos *mgo.Session
var g_ishk bool
var g_servaddr, g_mgourl, g_port string

func init() {
	// env
	g_mgourl = os.Getenv("MONGOLAB_URI")
	g_port = os.ExpandEnv("$PORT")
	// loc-or-hk
	g_ishk = g_mgourl != ""
	if g_ishk {
		g_servaddr = "http://127.0.0.1:8088"
	} else {
		g_mgourl = "127.0.0.1:27017/gospel"
		g_port = "8088"
		g_servaddr = HKADDRESS
	}
	fmt.Println("Running on/against " + g_servaddr)
}

func main() {
	// templates
	g_tmpls = template.Must(template.ParseFiles(tmplFName("default")))
	// db
	var err error
	g_mgos, err = mgo.Dial(g_mgourl)
	if err != nil {
		panic(err)
	}
	defer g_mgos.Close()
	// handles
	http.Handle("/sts/", http.FileServer(http.Dir("./res/")))
	http.HandleFunc("/msg/", msgHandler)
	http.HandleFunc("/", defaultHandler)
	// http server
	panic(http.ListenAndServe(":"+g_port, nil))
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

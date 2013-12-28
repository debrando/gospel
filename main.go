package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	TEXTHTML   = "text/html; charset=utf-8"
	APPJSON    = "application/json; charset=utf-8"
	LOCADDRESS = "http://127.0.0.1:8088"
	HKADDRESS  = "http://gospel99.herokuapp.com"
)

type BaseMsg struct {
	Success bool
	Message string
	Ts      time.Time
}

var (
	g_tmpls                      *template.Template
	g_mgos                       *mgo.Session
	g_servaddr, g_mgourl, g_port string
)

func init() {
	// env
	g_mgourl = os.Getenv("MONGOLAB_URI")
	g_port = os.Getenv("PORT")
	g_servaddr = os.Getenv("LOCALTEST")
	// loc-or-hk
	if g_mgourl == "" {
		g_mgourl = "127.0.0.1:27017/gospel"
		g_port = "8088"
		fmt.Println("Running on local ", LOCADDRESS)
	}
	if os.Getenv("LOCALTEST") == "y" {
		g_servaddr = LOCADDRESS
		fmt.Println("Tests to local ", LOCADDRESS)
	} else {
		g_servaddr = HKADDRESS
	}
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
	g_mgos.SetMode(mgo.Eventual, true)
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
	c := g_mgos.DB("").C("messages")
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", APPJSON)
		var v []BaseMsg
		err := c.Find(bson.M{"success": false}).Sort("ts").Limit(100).Iter().All(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		b, err := json.Marshal(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprint(w, string(b))
	case "POST":
		r.ParseForm()
		for _, m := range r.Form["msg"] {
			err := c.Insert(&BaseMsg{Success: false, Message: m, Ts: time.Now()})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				log.Println("Added message: ", m)
			}
		}
		if len(r.Form["msg"]) > 0 {
			http.Error(w, "Messages added", http.StatusNoContent)
		} else {
			http.Error(w, "No message given", http.StatusBadRequest)
		}
	default:
		http.Error(w, r.Method+" not allowed", http.StatusMethodNotAllowed)
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

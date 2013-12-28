package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/ugorji/go/codec"
	"html/template"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	TEXTHTML   = "text/html"
	APPJSON    = "application/json"
	MSGPACK    = "application/x-msgpack"
	LOCADDRESS = "http://127.0.0.1:8088"
	HKADDRESS  = "http://gospel99.herokuapp.com"
)

type BaseMsg struct {
	Success bool
	Message string
	Ts      time.Time
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
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
	if g_mgourl != "" && os.Getenv("LOCALTEST") != "n" {
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
	http.HandleFunc("/msg/", makeGzipHandler(msgHandler))
	http.HandleFunc("/", makeGzipHandler(defaultHandler))
	// http server
	panic(http.ListenAndServe(":"+g_port, nil))
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

// Wrapper for handling gzip encoding, https://gist.github.com/the42/1956518
func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "default")
}

// return if accept header has the given content type (or */*, or nothing)
func checkContent(ah string, ctype string) bool {
	m, err := regexp.MatchString(`(?i).*(\s+|^)(`+ctype+`|\*/\*)(;|$).*|^$`, ah)
	if err != nil {
		panic(err)
	}
	return m
}

// set the content type if originally accepted, returning if done
func setContent(w http.ResponseWriter, r *http.Request, ctype string) bool {
	isit := checkContent(r.Header.Get("Accept"), ctype)
	if isit {
		w.Header().Set("Content-Type", ctype)
	}
	return isit
}

// Message Handlers
func msgHandler(w http.ResponseWriter, r *http.Request) {
	c := g_mgos.DB("").C("messages")
	switch r.Method {
	case "GET":
		var v []BaseMsg
		err := c.Find(bson.M{"success": false}).Sort("ts").Limit(100).Iter().All(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if setContent(w, r, APPJSON) {
			b, err := json.Marshal(v)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			fmt.Fprint(w, string(b))
		} else if setContent(w, r, MSGPACK) {
			var mh codec.MsgpackHandle
			enc := codec.NewEncoder(w, &mh)
			err := enc.Encode(v)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Unsupported media type "+r.Header.Get("Accept"), http.StatusNotImplemented)
		}
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
	//log.Println(w.Header())
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

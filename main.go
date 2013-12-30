package main

import (
	"compress/gzip"
	"compress/zlib"
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
	"strconv"
	"strings"
	"time"
)

const (
	TEXTHTML   = "text/html"
	APPJSON    = "application/json"
	MSGPACK    = "application/x-msgpack"
	BSON       = "application/x-bson"
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
	http.HandleFunc("/msg/", makeGzipHandler(msgHandler, zlib.BestSpeed))
	http.HandleFunc("/", makeGzipHandler(defaultHandler, zlib.BestSpeed))
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
func makeGzipHandler(fn http.HandlerFunc, complvl int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ae := r.Header.Get("Accept-Encoding")
		if strings.Contains(ae, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			ow, _ := gzip.NewWriterLevel(w, complvl)
			defer ow.Close()
			gzr := gzipResponseWriter{Writer: ow, ResponseWriter: w}
			fn(gzr, r)
		} else if strings.Contains(ae, "deflate") {
			w.Header().Set("Content-Encoding", "deflate")
			ow, _ := zlib.NewWriterLevel(w, complvl)
			defer ow.Close()
			gzr := gzipResponseWriter{Writer: ow, ResponseWriter: w}
			fn(gzr, r)
		} else {
			fn(w, r)
			return
		}
	}
}

// Default Request Handler
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "default")
}

// return if accepts header has the given content type (or */*, or nothing)
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
	r.ParseForm()
	switch r.Method {
	case "GET":
		lim := 1000
		for _, m := range r.Form["limit"] {
			l, err := strconv.Atoi(m)
			if err == nil {
				lim = l
			}
		}
		q := bson.M{"success": false}
		var v []BaseMsg
		err := c.Find(q).Sort("ts").Limit(lim).All(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if setContent(w, r, APPJSON) {
			b, err := json.Marshal(v)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(b)
		} else if setContent(w, r, MSGPACK) {
			var mh codec.MsgpackHandle
			enc := codec.NewEncoder(w, &mh)
			err := enc.Encode(v)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if setContent(w, r, BSON) {
			b, err := bson.Marshal(v)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Write(b)
		} else {
			http.Error(w, "Unsupported media type "+r.Header.Get("Accept"), http.StatusNotImplemented)
		}
	case "POST":
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

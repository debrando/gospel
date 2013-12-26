// main_testing.go
package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

const TEXTHTML = "text/html; charset=utf-8"

var g_address string

func init() {
	// parameters
	flag.StringVar(&g_address, "g_address", "127.0.0.1:8088", "Server address")
	flag.Parse()
}

func RestGet(tb testing.TB, res string, ctype string) []byte {
	resp, err := http.Get("http://" + g_address + "/")
	if err != nil {
		tb.Error(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		tb.Error("Wrong response code ", err, ", expecting 200")
		return nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tb.Error(err)
		return nil
	}
	if resp.Header["Content-Type"][0] != ctype {
		tb.Error("Wrong Content-Type", resp.Header["Content-Type"][0])
	}
	return d
}

func TestHome(t *testing.T) {
	data := RestGet(t, "/", TEXTHTML)
	if data == nil {
		t.Error("No data found")
		return
	}
	reTa := regexp.MustCompile(`<\w+>`)
	allT := reTa.FindAll(data, 5)
	if allT == nil {
		t.Error("Haven't found any tag on body")
		return
	}
	for _, tag := range allT {
		t.Log("Found: ", string(tag))
	}
}

func BenchmarkHome(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/", TEXTHTML)
	}
}

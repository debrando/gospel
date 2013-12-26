// main_testing.go
package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

var g_address string

func init() {
	// parameters
	flag.StringVar(&g_address, "g_address", "127.0.0.1:8088", "Server address")
	flag.Parse()
}

func RestGet(tb testing.TB, res string) []byte {
	resp, err := http.Get("http://" + g_address + "/")
	if err != nil {
		tb.Error(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		tb.Error("Wrong response code ", err, ", expecting 200")
		return nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tb.Error(err)
		return nil
	}
	return d
}

func TBHome(tb testing.TB) string {
	resp, err := http.Get(g_address)
	if err != nil {
		tb.Error(err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		tb.Error("Wrong response code ", err, ", expecting 200")
		return ""
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tb.Error(err)
		return ""
	}
	return string(d)
}

func TestHome(t *testing.T) {
	data := RestGet(t, "/")
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
		RestGet(b, "/")
	}
}

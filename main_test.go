// main_testing.go
package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func TestHome(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:8088")
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	reTa := regexp.MustCompile(`<\w+>`)
	allT := reTa.FindAll(d, 5)
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
		resp, err := http.Get("http://127.0.0.1:8088")
		if err != nil {
			b.Error(err)
			return
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			b.Error(err)
		}
	}
}

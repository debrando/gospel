// main_testing.go
package main

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func TBHome(tb testing.TB) string {
	resp, err := http.Get("http://127.0.0.1:8088")
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
	data := TBHome(t)
	if data == "" {
		t.Error("No data found")
		return
	}
	reTa := regexp.MustCompile(`<\w+>`)
	allT := reTa.FindAllString(data, 5)
	if allT == nil {
		t.Error("Haven't found any tag on body")
		return
	}
	for _, tag := range allT {
		t.Log("Found: ", tag)
	}
}

func BenchmarkHome(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TBHome(b)
	}
}

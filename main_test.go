// main_testing.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func RestGet(tb testing.TB, res string, ctype string) []byte {
	resp, err := http.Get(g_servaddr + res)
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
		return nil
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

func TestGetMsgs(t *testing.T) {
	data := RestGet(t, "/msg/", APPJSON)
	if data == nil {
		t.Error("No data found")
		return
	}
	var msgs []BaseMsg
	err := json.Unmarshal(data, &msgs)
	if err != nil {
		t.Error(err, " on blob ", string(data))
	}
	for _, msg := range msgs {
		if !msg.Success {
			t.Log(msg.Message)
		} else {
			t.Error("Wrong message ", msg)
		}
	}
}

func BenchmarkGetMsgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/msg/", APPJSON)
	}
}

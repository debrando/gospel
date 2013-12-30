// main_testing.go
package main

import (
	"bytes"
	"encoding/json"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"
)

func RestGet(tb testing.TB, res string, ctype string, gzip bool) []byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", g_servaddr+res, nil)
	if !gzip {
		req.Header.Add("Accept-Encoding", "")
	}
	req.Header.Add("Accept", ctype)
	resp, err := client.Do(req)
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
	if !checkContent(resp.Header.Get("Content-Type"), ctype) {
		tb.Error("Wrong Content-Type", resp.Header.Get("Content-Type"))
	}
	return d
}

// UTs

func TestHome(t *testing.T) {
	data := RestGet(t, "/", TEXTHTML, true)
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

func TestGetMsgsJSON(t *testing.T) {
	data := RestGet(t, "/msg/", APPJSON, true)
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

func TestGetMsgsMSGPack(t *testing.T) {
	r := bytes.NewBuffer(RestGet(t, "/msg/", MSGPACK, true))
	if r == nil {
		t.Error("No data found")
		return
	}
	var msgs []BaseMsg
	var mh codec.MsgpackHandle
	dec := codec.NewDecoder(r, &mh)
	err := dec.Decode(&msgs)
	if err != nil {
		t.Error(err, " on blob msgpacked")
	}
	for _, msg := range msgs {
		if !msg.Success {
			t.Log(msg.Message)
		} else {
			t.Error("Wrong message ", msg)
		}
	}
}

// Benchmarks

func BenchmarkHomePlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/", TEXTHTML, false)
	}
}

func BenchmarkHomeGZip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/", TEXTHTML, true)
	}
}

func BenchmarkGetMsgsJSonPlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/msg/", APPJSON, false)
	}
}
func BenchmarkGetMsgsJSonGZip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/msg/", APPJSON, true)
	}
}

func BenchmarkGetMsgsMSGPackPlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/msg/", MSGPACK, false)
	}
}

func BenchmarkGetMsgsMSGPackGZip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RestGet(b, "/msg/", MSGPACK, true)
	}
}

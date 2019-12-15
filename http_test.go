package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/vodafon/verlog"
)

func TestProcessRequest(t *testing.T) {
	log = verlog.New(0)
	var tests = []struct {
		word   string
		method string
		exp    string
		status int
	}{
		{"api", "GET", "GET http://service.com/target-api\n", 200},
		{"comp", "GET", "", 302},
		{"graphql", "POST", "POST http://service.com/target-graphql\n", 200},
	}
	for _, tt := range tests {
		w := bytes.NewBuffer([]byte{})
		cl := client{
			c: newTestClient(tt.status),
			w: w,
		}
		service := Service{
			Name:   "test",
			Method: tt.method,
			URL:    "http://service.com/COMPANY",
			Analysis: &Analysis{
				Status: 200,
			},
		}
		request := requestService(service, "target", "-", tt.word)
		processRequest(request, cl)
		res := w.String()
		if tt.exp != res {
			t.Errorf("Incorrect result. Expected %q, got %q\n", tt.exp, res)
		}
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(status int) *http.Client {
	return &http.Client{
		Transport: newRoundTrip(status),
	}
}

func newRoundTrip(status int) roundTripFunc {
	return func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: status,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	}
}

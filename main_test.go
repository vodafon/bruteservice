package main

import (
	"strings"
	"testing"
)

func TestPermutations(t *testing.T) {
	requestC := make(chan Request, 5)
	res := ""
	service := Service{
		Name:   "test",
		Method: "GET",
		URL:    "http://s.com/COMPANY",
		Analysis: &Analysis{
			Status: 200,
		},
	}
	*flagCompany = "target"
	permutations(requestC, []Service{service}, "api")
	close(requestC)
	for rr := range requestC {
		res += strings.TrimPrefix(rr.URL, "http://s.com/") + ";"
	}
	exp := "targetapi;apitarget;api-target;target-api;"
	if exp != res {
		t.Errorf("Incorrect result. Expected %q, got %q\n", exp, res)
	}
}

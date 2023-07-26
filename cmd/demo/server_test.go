package main

import (
	"io"
	"net/http"
	"testing"
)

const (
	loadBalancer             = "http://localhost:8080"
	server1                  = "http://localhost:8081"
	server2                  = "http://localhost:8082"
	numReqRoundRobin         = 10
	numReqWeightedRoundRobin = 15
	server1OkResp            = "All good from server 8081."
	server2OkResp            = "All good from server 8082."
)

func TestRoundRobin(t *testing.T) {
	for i := 0; i < numReqRoundRobin; i++ {
		t.Logf("Sending req #%d ...", i+1)

		res, err := http.Get(loadBalancer)
		if err != nil {
			t.Fatalf("Error sending the request: %v", err)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading the response body: %v", err)
		}

		gotBody := string(body)
		var wantBody string

		if i%2 == 0 {
			wantBody = server1OkResp
		} else {
			wantBody = server2OkResp
		}

		if gotBody != wantBody {
			t.Errorf("RoundRobin body = %q; want %q", gotBody, wantBody)
		}
	}
}

func TestWeightedRoundRobin(t *testing.T) {
	for i := 0; i < numReqWeightedRoundRobin; i++ {
		t.Logf("Sending req #%d ...", i+1)

		res, err := http.Get(loadBalancer)
		if err != nil {
			t.Fatalf("Error sending the request: %v", err)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading the response body: %v", err)
		}

		gotBody := string(body)
		var wantBody string

		t.Log("response: ", string(body))

		if i < 10 {
			wantBody = server1OkResp
		} else {
			wantBody = server2OkResp
		}

		if gotBody != wantBody {
			t.Errorf("WeightedRoundRobin body = %q; want %q", gotBody, wantBody)
		}
	}
}

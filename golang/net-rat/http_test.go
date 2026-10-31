//        ____()()     NetRat v0.2.4
//       /      @@     ~~~~~~~~~~~~~
// `~~~~~\_;m__m._>o   A tiny Go experiment
//
// Author: Giovanni Squillero <giovanni.squillero@polito.it>
// Copyright © 2024-26 Giovanni Squillero / Politecnico di Torino
// https://github.com/squillero/programmer-zendo
// Free under certain conditions — see the license for details.

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchRawSendsHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("Accept"))
	}))
	defer srv.Close()

	raw := FetchRaw(map[string]string{"Accept": "application/rdap+json"}, context.Background(), srv.URL)
	if string(raw) != "application/rdap+json" {
		t.Errorf("FetchRaw() = %q, want the Accept header echoed back", raw)
	}
}

func TestFetchRawNotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	if raw := FetchRaw(nil, context.Background(), srv.URL); raw != nil {
		t.Errorf("FetchRaw() = %q, want nil", raw)
	}
}

func TestFetchRawCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "1.2.3.4")
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if raw := FetchRaw(nil, ctx, srv.URL); raw != nil {
		t.Errorf("FetchRaw() = %q, want nil with a canceled context", raw)
	}
}

func TestFetchRawBadURL(t *testing.T) {
	if raw := FetchRaw(nil, context.Background(), "://bad"); raw != nil {
		t.Errorf("FetchRaw() = %q, want nil", raw)
	}
}

func TestFetchHttp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, "  1.2.3.4\n")
	}))
	defer srv.Close()

	if got := FetchHttp(context.Background(), srv.URL); got != "1.2.3.4" {
		t.Errorf("FetchHttp() = %q, want %q", got, "1.2.3.4")
	}
	if got := FetchHttp(context.Background(), srv.URL+"/missing"); got != "" {
		t.Errorf("FetchHttp(missing) = %q, want \"\"", got)
	}
}

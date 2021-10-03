package main

import (
	"fmt"
	"time"
	"testing"
	"net/http"
	"net/url"
	"strings"
	"net/http/httptest"
	"hashbrown/datastore"
)

func TestStats(t *testing.T) {
	srv := NewServer("")
	srv.db = datastore.NewHashStore()

	r := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("TestStats expected a 200 status code, but received %d", w.Code)
	}
	fmt.Printf("body: %s\n", w.Body)

	r = httptest.NewRequest("GET", "/status", nil)
	w = httptest.NewRecorder()

	srv.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("TestStats expected a 404 status code, but received %d", w.Code)
	}
}

func TestHash(t *testing.T) {
	srv := NewServer("")
	srv.db = datastore.NewHashStore()
	
	pw := "abc123"
	dataMap := url.Values{"password": []string{pw}}
	formData := strings.NewReader(dataMap.Encode())
	req := httptest.NewRequest("POST", "/hash", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)

	req = httptest.NewRequest("GET", "/stats", nil)
	w = httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}

	//time.Sleep(5*time.Second+1000)
	time.Sleep(1000)

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}
}

func BenchmarkHash(t *testing.B) {
	srv := NewServer("")
	srv.db = datastore.NewHashStore()
	
	for i:=0; i<1000000; i++ {
		pw := fmt.Sprintf("abcpw%d", i)
		dataMap := url.Values{"password": []string{pw}}
		formData := strings.NewReader(dataMap.Encode())
		req := httptest.NewRequest("POST", "/hash", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		// fmt.Printf("body: %s\n", w.Body)
	}

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}
/*
	time.Sleep(5*time.Second+1000)

	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}
*/
}
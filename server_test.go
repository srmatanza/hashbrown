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
	srv.db = datastore.NewPoolStore()

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

func TestShutdown(t *testing.T) {
	srv := NewServer("")
	srv.db = datastore.NewSyncStore()

	req := httptest.NewRequest("POST", "/shutdown", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	srv.WaitForQuit()
	fmt.Printf("body: %s\n", w.Body)
}

func TestHash(t *testing.T) {
	srv := NewServer("")
	srv.db = datastore.NewPoolStore()

	count := 10
	
	for i:=0; i<count; i++ {
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

	time.Sleep(5*time.Second+1000)
	
	for i:=0; i<count; i++ {
		strHashId := fmt.Sprintf("/hash/%d", i+1)
		req = httptest.NewRequest("GET", strHashId, nil)
		w = httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		fmt.Printf("body: %s\n", w.Body)
	}

	req = httptest.NewRequest("GET", "/stats", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}
}

func BenchmarkPoolStore(b *testing.B) {
	genHashes(500000, datastore.NewPoolStore(), b)
}

func BenchmarkSyncStore(b *testing.B) {
	genHashes(500000, datastore.NewSyncStore(), b)
}

func genHashes(count int, db Datastore, b *testing.B) {
	srv := NewServer("")

	srv.db = db
	
	for i:=0; i<count; i++ {
		pw := fmt.Sprintf("abcpw%d", i)
		dataMap := url.Values{"password": []string{pw}}
		formData := strings.NewReader(dataMap.Encode())
		req := httptest.NewRequest("POST", "/hash", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		// fmt.Printf("body: %s\n", w.Body)
	}

	// time.Sleep(5*time.Second+100)
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)
	fmt.Printf("body: %s\n", w.Body)
	if w.Code != http.StatusOK {
		b.Errorf("expected status code %d but received %d", http.StatusOK, w.Code)
	}
}
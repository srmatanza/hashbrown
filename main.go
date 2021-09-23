package main

import (
  "fmt"
  "flag"
  "log"
  "net/http"
  "regexp"
  "strconv"
)

var addr = flag.String("addr", ":8123", "hashbrown service address")

var validHashId = regexp.MustCompile(`^/hash/([1-9]+[0-9]*)$`)

func main() {
  flag.Parse()

  quit := make(chan bool)

  http.Handle("/stats", http.HandlerFunc(statsHandler))
  http.Handle("/hash/", http.HandlerFunc(getHashHandler))
  http.Handle("/hash", http.HandlerFunc(postHashHandler))
  http.Handle("/shutdown", http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
    shutdownHandler(w, req, quit)
  }))

  go Serve()
  <-quit
  log.Printf("Closing the server gracefully")
}

func Serve() {
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

func getHashHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    if hashId := validHashId.FindStringSubmatch(req.URL.Path); hashId != nil {
      if uintHashId, err := strconv.ParseUint(hashId[1], 10, 32); err == nil {
        fmt.Fprintf(w, "getting hash for id: %d", uintHashId)
        return
      }
    }
  }
  http.Error(w, "Bad Request", 400)
}

func postHashHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "POST" {
    if payload := req.PostFormValue("password"); payload != "" {
      fmt.Fprintf(w, "posting hash for: %q", payload)
      return
    }
  }
  http.Error(w, "Bad Request", 400)
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    fmt.Fprintf(w, "Here's yer stats. %q", req.Method)
    return
  }
  http.Error(w, "Bad Request", 400)
}

func shutdownHandler(w http.ResponseWriter, req *http.Request, quit chan bool) {
  if req.Method == "POST" {
    fmt.Fprintf(w, "Shutting down the server.")
    quit <- true
    return
  }
  http.Error(w, "Bad Request", 400)
}


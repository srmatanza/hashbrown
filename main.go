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
        if entry, ok := getHash(uint64(uintHashId)); ok == true {
          b64hash := encodeHash(entry.hash)
          fmt.Fprintf(w, "getting hash for id: %d, %q", uintHashId, b64hash)
          return
        }
        http.Error(w, "Hash Not Found", 404)
        return
      }
    }
  }
  http.Error(w, "Bad Request", 400)
}

func postHashHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "POST" {
    if payload := req.PostFormValue("password"); payload != "" {
      computedHash := computeHash(payload)
      hashId := storeHash(computedHash)
      fmt.Fprintf(w, "posting hash: %d", hashId)
      return
    }
  }
  http.Error(w, "Bad Request", 400)
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    w.Header().Add("Content-Type", "application/json")
    fmt.Fprintf(w, "%q", getStatsJSON())
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


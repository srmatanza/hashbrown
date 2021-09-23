package main

import (
  "fmt"
  "flag"
  "log"
  "net/http"
  "regexp"
  "strconv"
  "hashbrown/datastore"
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
        if entry, ok := datastore.GetHash(uint64(uintHashId)); ok == true {
          b64hash := encodeHash(entry.Hash)
          w.Header().Add("Content-Type", "application/json")
          fmt.Fprintf(w, "{\"hashId\": %d, \"hash\": %q}\n", uintHashId, b64hash)
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
      hashId := datastore.PutHash(computedHash)

      w.Header().Add("Content-Type", "application/json")
      fmt.Fprintf(w, "{\"hashId\": %d}\n", hashId)
      return
    }
  }
  http.Error(w, "Bad Request", 400)
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    w.Header().Add("Content-Type", "application/json")
    total, avg := datastore.GetStats()
    fmt.Fprintf(w, "{\"total\": %d,\"average\": %d}\n", total, avg)
    return
  }
  http.Error(w, "Bad Request", 400)
}

func shutdownHandler(w http.ResponseWriter, req *http.Request, quit chan bool) {
  if req.Method == "POST" {
    fmt.Fprintf(w, "Shutting down the server.\n")
    quit <- true
    return
  }
  http.Error(w, "Bad Request", 400)
}


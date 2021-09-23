package main

import (
  "fmt"
  "time"
  "regexp"
  "strconv"
  "net/http"
  "hashbrown/datastore"
)

var validHashId = regexp.MustCompile(`^/hash/([1-9]+[0-9]*)$`)

func getHashHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    if hashId := validHashId.FindStringSubmatch(req.URL.Path); hashId != nil {
      if uintHashId, err := strconv.ParseUint(hashId[1], 10, 32); err == nil {
        if entry, ok := datastore.GetHash(uint64(uintHashId)); ok == true {
          b64hash := encodeHash(entry.Hash)
          w.Header().Add("Content-Type", "application/json")
          fmt.Fprintf(w, "{\"id\": %d, \"hash\": %q}\n", uintHashId, b64hash)
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

      w := w
      go func() {
        start := time.Now()
        computedHash := computeHash(payload)
        tdelta := time.Now().Sub(start)

        hashId := datastore.PutHash(computedHash, tdelta)

        w.Header().Add("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"id\": %d}\n", hashId)
      }()
      return
    }
  }
  http.Error(w, "Bad Request", 400)
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
  if req.Method == "GET" {
    w.Header().Add("Content-Type", "application/json")
    total, avg := datastore.GetStats()
    fmt.Fprintf(w, "{\"total\": %d,\"average\": %d}\n", total, avg/1000)
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


package main

import (
  "fmt"
  "strconv"
  "net/http"
  "encoding/base64"
)

func (s *server) handleStatsGet() http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {
    w.Header().Add("Content-Type", "application/json")
    total, avg := s.db.GetStats()
    fmt.Fprintf(w, "{\"total\": %d,\"average\": %d}\n", total, avg/1000)
    return
  }
}

func (s *server) handleHashGet() http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {
    w.Header().Add("Content-Type", "application/json")
    pathParams := req.Context().Value(ctxPathParams{}).([]string)
    if uintHashId, err := strconv.ParseUint(pathParams[0], 10, 32); err == nil {
      if entry, ok := s.db.GetHash(uintHashId); ok {
        b64hash := base64.StdEncoding.EncodeToString(entry.Hash)
        w.Header().Add("Content-Type", "application/json")
        fmt.Fprintf(w, "{\"id\": %d, \"hash\": %q}\n", uintHashId, b64hash)
        return
      }
    }
    http.NotFound(w, req)
  }
}

func (s *server) handleHashPost() http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {
    if payload := req.PostFormValue("password"); payload != "" {
      hashId := s.db.PutHash(payload)
      w.Header().Add("Content-Type", "application/json")
      fmt.Fprintf(w, "{\"id\": %d}\n", hashId)
      return
    }
    http.Error(w, "Bad Request", 400)
  }
}

func (s *server) handleShutdownPost() http.HandlerFunc {
  return func(w http.ResponseWriter, req *http.Request) {
    fmt.Fprintf(w, "Shutting down the server.\n")
    s.shutdown()
  }
}
